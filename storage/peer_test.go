package storage

import (
	"bytes"
	"context"
	"github.com/amirylm/libp2p-facade/core"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/importer/balanced"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestStorageNodeOffline(t *testing.T) {
	psk := core.PNetSecret()

	base1 := core.NewBasePeer(context.Background(), core.NewConfig(nil, psk, nil))
	n1 := NewStoragePeer(base1, true)
	defer n1.Close()

	plaintext := "Some data... more or less"
	root, err := add(n1, plaintext)
	assert.Nil(t, err)

	b, err := GetBytes(n1, root.Cid())
	assert.Nil(t, err)
	assert.Equal(t, plaintext, string(b))
}

func TestStorageNode(t *testing.T) {
	psk := core.PNetSecret()
	nodes, err := core.SetupGroup(3, func(onPeerFound core.OnPeerFound) core.LibP2PPeer {
		return newStoragePeer(psk, onPeerFound)
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(nodes))

	time.Sleep(time.Millisecond * 500)

	// add value in first node
	n1 := nodes[0].(StoragePeer)
	plaintext := "Some data... more or less"
	root, err := add(n1, plaintext)
	assert.Nil(t, err)

	// get value from second node
	n2 := nodes[1].(StoragePeer)
	b, err := GetBytes(n2, root.Cid())
	assert.Nil(t, err)
	assert.Equal(t, plaintext, string(b))
}

func add(n StoragePeer, data string) (ipld.Node, error) {
	r := bytes.NewReader([]byte(data))
	cb, err := cidBuilder("")
	if err != nil {
		return nil, err
	}
	return Add(n, []byte{}, r, cb, balanced.Layout)
}

func newStoragePeer(psk pnet.PSK, onPeerFound core.OnPeerFound) StoragePeer {
	cfg := core.NewConfig(nil, psk, nil)
	cfg.Discovery = core.NewDiscoveryConfig(onPeerFound)
	base := core.NewBasePeer(context.Background(), cfg)
	peer := NewStoragePeer(base, false)
	peer.Logger().Infof("new peer: %s", peer.Host().ID().Pretty())
	return peer
}
