package core

import (
	"bytes"
	"context"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestPubSubEmitter(t *testing.T) {
	n := 4
	psk := PNetSecret()
	nodes, err := SetupGroup(n, func(onPeerFound OnPeerFound) LibP2PPeer {
		node := newPubSubPeer(psk, onPeerFound)
		return node
	})
	assert.Nil(t, err)
	if nodes == nil {
		assert.FailNow(t, "could not setup nodes")
	}
	assert.Equal(t, n, len(nodes))

	var pswg sync.WaitGroup
	data := []byte("data:my-topic")
	pswg.Add(1)
	sub1, err := Subscribe(nodes[0], "my-topic")
	assert.Nil(t, err)
	go func() {
		for {
			msg, err := sub1.Next(nodes[0].Context())
			assert.Nil(t, err)
			assert.True(t, bytes.Equal(data, msg.Data))
			pswg.Done()
		}
	}()

	sub2, err := Subscribe(nodes[1],"other-topic")
	assert.Nil(t, err)
	go func() {
		for {
			sub2.Next(nodes[1].Context())
			assert.Fail(t, "should not receive a message")
			return
		}
	}()

	topic3, err := Topic(nodes[2], "my-topic")
	assert.Nil(t, err)
	go func() {
		time.Sleep(1000 * time.Millisecond)
		topic3.Publish(nodes[2].Context(), data[:])
	}()
	pswg.Wait()
}

func newPubSubPeer(psk pnet.PSK, onPeerFound OnPeerFound) *BasePeer {
	cfg := NewConfig(nil, psk, nil)
	cfg.Discovery = NewDiscoveryConfig(onPeerFound)
	n := NewBasePeer(context.Background(), cfg)
	n.Logger().Infof("new peer: %s", n.Host().ID().Pretty())
	return n
}

