package storage

import (
	"bytes"
	"context"
	"github.com/amirylm/libp2p-facade/core"
	ds "github.com/ipfs/go-datastore"
	crdt "github.com/ipfs/go-ds-crdt"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIpldNodeWithCrdt(t *testing.T) {
	t.SkipNow()

	crdts := []*crdt.Datastore{}
	topicName := "crdt-test"
	psk := core.PNetSecret()

	_, err := core.SetupGroup(3, func(onPeerFound core.OnPeerFound) core.LibP2PPeer {
		node, c, err := newCrdtNode(psk, onPeerFound, topicName)
		assert.Nil(t, err)
		crdts = append(crdts, c)
		return node
	})
	assert.Nil(t, err)

	time.Sleep(time.Millisecond * 500)

	// add value in first node
	state1 := []byte(`{"state": "val1"}`)
	state11 := []byte(`{"state": "val11"}`)
	//state111 := []byte(`{"state": "val111"}`)
	state2 := []byte(`{"state": "val2"}`)
	state22 := []byte(`{"state": "val22"}`)
	k := "state-key"
	err = crdts[0].Put(ds.NewKey(k), state1)
	assert.Nil(t, err)

	err = crdts[1].Put(ds.NewKey(k), state2)
	assert.Nil(t, err)

	err = crdts[0].Put(ds.NewKey(k), state11)
	assert.Nil(t, err)

	//err = crdt1.Put(ds.NewKey(k), state111)
	//assert.Nil(t, err)

	err = crdts[1].Put(ds.NewKey(k), state22)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	val1, err := crdts[0].Get(ds.NewKey(k))
	assert.Nil(t, err)

	val2, err := crdts[1].Get(ds.NewKey(k))
	assert.Nil(t, err)

	val3, err := crdts[2].Get(ds.NewKey(k))
	assert.Nil(t, err)

	assert.True(t, bytes.Equal(val1, val2))
	assert.True(t, bytes.Equal(val2, val3))
	assert.True(t, bytes.Equal(val1, state22))
}

func newCrdtNode(psk pnet.PSK, onPeerFound core.OnPeerFound, crdtTopic string) (StoragePeer, *crdt.Datastore, error) {
	cfg := core.NewConfig(nil, psk, nil)
	cfg.Discovery = core.NewDiscoveryConfig(onPeerFound)
	base := core.NewBasePeer(context.Background(), cfg)
	node := NewStoragePeer(base, false)
	node.Logger().Infof("new node: %s", node.Host().ID().Pretty())
	c, err := ConfigureCrdt(node, crdtTopic, nil)
	if err != nil {
		return nil, nil, err
	}
	go core.AutoClose(node.Context(), c)
	return node, c, nil
}
