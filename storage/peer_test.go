package storage

import (
	"bytes"
	"context"
	"github.com/amirylm/libp2p-facade/core"
	ipld "github.com/ipfs/go-ipld-format"
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
	n := 4
	psk := core.PNetSecret()
	nodes, err := core.SetupGroup(n, func() core.LibP2PPeer {
		return newStoragePeer(psk)
	})
	assert.Nil(t, err)
	assert.Equal(t, n, len(nodes))

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

func TestAddDir(t *testing.T) {
	n := 4
	psk := core.PNetSecret()
	nodes, err := core.SetupGroup(n, func() core.LibP2PPeer {
		return newStoragePeer(psk)
	})
	assert.Nil(t, err)
	assert.Equal(t, n, len(nodes))

	time.Sleep(time.Millisecond * 500)

	n0 := nodes[0].(StoragePeer)
	dir, dirnode, err := AddDir(n0)
	if err != nil {
		t.Fatal(err)
	}
	n0.Logger().Infof("dir created %s", dirnode.Cid().String())

	plaintext := "Child data..."
	name := "somedata"
	datanode, err := add(n0, "Child data...")
	assert.Nil(t, err)
	dir, dirnode, err = AddToDir(n0, dir, name, datanode)
	assert.Nil(t, err)
	n0.Logger().Infof("data %s added to %s", datanode.Cid().String(), dirnode.Cid().String())

	n1 := nodes[1].(StoragePeer)
	dirN1, err := LoadDir(n1, dirnode.Cid())
	assert.Nil(t, err)
	datanodeN1, err := dirN1.Find(n1.Context(), name)
	b, err := GetBytes(n1, datanodeN1.Cid())
	assert.Nil(t, err)
	assert.Equal(t, plaintext, string(b))
}

func add(n StoragePeer, data string) (ipld.Node, error) {
	r := bytes.NewReader([]byte(data))
	cb, err := NewCidBuilder("")
	if err != nil {
		return nil, err
	}
	return Add(n, r, cb, nil)
}

func newStoragePeer(psk pnet.PSK) StoragePeer {
	cfg := core.NewConfig(nil, psk, nil)
	base := core.NewBasePeer(context.Background(), cfg)
	peer := NewStoragePeer(base, false)
	peer.Logger().Infof("new peer: %s", peer.Host().ID().Pretty())
	return peer
}
