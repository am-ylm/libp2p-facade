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

func TestPeerWithCrdt(t *testing.T) {
	n := 4
	crdts := []*crdt.Datastore{}
	topicName := "crdt-test"
	psk := core.PNetSecret()

	_, err := core.SetupGroup(n, func() core.LibP2PPeer {
		node, c, err := newCrdtPeer(psk, topicName)
		assert.Nil(t, err)
		crdts = append(crdts, c)
		return node
	})
	assert.Nil(t, err)

	time.Sleep(time.Millisecond * 500)

	// manipulating values from multiple peers
	state1 := []byte(`{"state": "val1"}`)
	state11 := []byte(`{"state": "val11"}`)
	state2 := []byte(`{"state": "val2"}`)
	state22 := []byte(`{"state": "val22"}`)
	k := "state-key"
	k2 := "state-key2"
	err = crdts[0].Put(ds.NewKey(k), state1)
	assert.Nil(t, err)

	err = crdts[1].Put(ds.NewKey(k), state2)
	assert.Nil(t, err)
	err = crdts[1].Put(ds.NewKey(k2), state2)
	assert.Nil(t, err)

	err = crdts[0].Put(ds.NewKey(k), state11)
	assert.Nil(t, err)

	err = crdts[1].Put(ds.NewKey(k), state22)
	assert.Nil(t, err)

	err = crdts[0].Put(ds.NewKey(k2), state11)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	val1, err := crdts[0].Get(ds.NewKey(k))
	assert.Nil(t, err)

	val2, err := crdts[1].Get(ds.NewKey(k))
	assert.Nil(t, err)
	k2val2, err := crdts[1].Get(ds.NewKey(k2))
	assert.Nil(t, err)

	val3, err := crdts[2].Get(ds.NewKey(k))
	assert.Nil(t, err)

	assert.True(t, bytes.Equal(val1, val2))
	assert.True(t, bytes.Equal(val2, val3))
	assert.True(t, bytes.Equal(val1, state22))
	assert.True(t, bytes.Equal(k2val2, state11))
}

func newCrdtPeer(psk pnet.PSK, crdtTopic string) (StoragePeer, *crdt.Datastore, error) {
	cfg := core.NewConfig(nil, psk, nil)
	base := core.NewBasePeer(context.Background(), cfg)
	peer := NewStoragePeer(base, false)
	peer.Logger().Infof("new peer: %s", peer.Host().ID().Pretty())
	c, err := ConfigureCrdt(peer, crdtTopic, nil)
	if err != nil {
		return nil, nil, err
	}
	go core.AutoClose(peer.Context(), c)
	return peer, c, nil
}
