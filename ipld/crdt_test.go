package ipld

//import (
//	"bytes"
//	"context"
//	"github.com/amirylm/priv-libp2p-node/core"
//	ds "github.com/ipfs/go-datastore"
//	"github.com/libp2p/go-libp2p-core/peer"
//	"github.com/stretchr/testify/assert"
//	"sync"
//	"testing"
//	"time"
//)
//
//func TestIpldNodeWithCrdt(t *testing.T) {
//	topicName := "crdt-test"
//	var discwg sync.WaitGroup
//	discwg.Add(1)
//
//	onPeerFound := core.OnPeerFoundWaitGroup(&discwg)
//	psk := core.PNetSecret()
//
//	base1 := core.NewBaseNode(context.Background(), core.NewConfig(nil, psk, nil), core.NewDiscoveryConfig(onPeerFound))
//	n1 := NewIpldNode(base1, false)
//	defer n1.Close()
//	crdt1, err := ConfigureCrdt(n1, nil, topicName)
//	n1.DHT().Bootstrap(n1.Context())
//	assert.Nil(t, err)
//	defer crdt1.Close()
//
//	time.Sleep(time.Second)
//
//	base2 := core.NewBaseNode(context.Background(), core.NewConfig(nil, psk, nil), core.NewDiscoveryConfig(onPeerFound))
//	n2 := NewIpldNode(base2, false)
//	defer n2.Close()
//	crdt2, err := ConfigureCrdt(n2, nil, topicName)
//	assert.Nil(t, err)
//	defer crdt2.Close()
//	core.Connect(n2, []peer.AddrInfo{{n1.Host().ID(), n1.Host().Addrs()}}, true)
//
//	discwg.Wait()
//
//	time.Sleep(time.Second * 2)
//
//	// add value in first node
//	state1 := []byte(`{"state": "val1"}`)
//	state11 := []byte(`{"state": "val11"}`)
//	//state111 := []byte(`{"state": "val111"}`)
//	state2 := []byte(`{"state": "val2"}`)
//	state22 := []byte(`{"state": "val22"}`)
//	k := "data-key"
//	err = crdt1.Put(ds.NewKey(k), state1)
//	assert.Nil(t, err)
//
//	err = crdt2.Put(ds.NewKey(k), state2)
//	assert.Nil(t, err)
//
//	err = crdt1.Put(ds.NewKey(k), state11)
//	assert.Nil(t, err)
//
//	//err = crdt1.Put(ds.NewKey(k), state111)
//	//assert.Nil(t, err)
//
//	err = crdt2.Put(ds.NewKey(k), state22)
//	assert.Nil(t, err)
//
//	time.Sleep(time.Second)
//
//	val1, err := crdt1.Get(ds.NewKey(k))
//	assert.Nil(t, err)
//
//	val2, err := crdt2.Get(ds.NewKey(k))
//	assert.Nil(t, err)
//
//	assert.True(t, bytes.Equal(val1, val2))
//	assert.True(t, bytes.Equal(val1, state22))
//}