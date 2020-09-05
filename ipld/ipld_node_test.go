package ipld

import (
	"bytes"
	"context"
	"github.com/amirylm/priv-libp2p-node/core"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/importer/balanced"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestIpldNodeOffline(t *testing.T) {
	psk := core.PNetSecret()

	base1 := core.NewBaseNode(context.Background(), core.NewConfig(nil, psk, nil), core.NewDiscoveryConfig(nil))
	n1 := NewIpldNode(base1, true)
	defer n1.Close()

	plaintext := "Some data... more or less"
	root, err := add(n1, plaintext)
	assert.Nil(t, err)

	b, err := GetBytes(n1, root.Cid())
	assert.Nil(t, err)
	assert.Equal(t, plaintext, string(b))
}

func TestIpldNode(t *testing.T) {
	var discwg sync.WaitGroup
	discwg.Add(2)

	onPeerFound := core.OnPeerFoundWaitGroup(&discwg)
	psk := core.PNetSecret()

	base1 := core.NewBaseNode(context.Background(), core.NewConfig(nil, psk, nil), core.NewDiscoveryConfig(onPeerFound))
	n1 := NewIpldNode(base1, false)
	defer n1.Close()
	n1.DHT().Bootstrap(n1.Context())

	time.Sleep(time.Second)

	base2 := core.NewBaseNode(context.Background(), core.NewConfig(nil, psk, nil), core.NewDiscoveryConfig(onPeerFound))
	n2 := NewIpldNode(base2, false)
	defer n2.Close()
	core.Connect(n2, []peer.AddrInfo{{n1.Host().ID(), n1.Host().Addrs()}}, true)

	discwg.Wait()

	// add value in first node
	plaintext := "Some data... more or less"
	root, err := add(n1, plaintext)
	assert.Nil(t, err)

	// get value from second node
	b, err := GetBytes(n2, root.Cid())
	assert.Nil(t, err)
	assert.Equal(t, plaintext, string(b))
}

func add(n IpldNode, data string) (ipld.Node, error) {
	r := bytes.NewReader([]byte(data))
	cb, err := cidBuilder("")
	if err != nil {
		return nil, err
	}
	return Add(n, []byte{}, r, cb, balanced.Layout)
}