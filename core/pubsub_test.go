package core

import (
	"bytes"
	"context"
	"errors"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/stretchr/testify/assert"
	"log"
	"sync"
	"testing"
	"time"
)

func TestPubSubEmitter(t *testing.T) {
	n := 3
	psk := PNetSecret()
	nodes, err := setupNodesGroup(n, psk)
	assert.Nil(t, err)
	if nodes == nil {
		assert.Fail(t, "could not setup nodes")
	}
	defer func() {
		for _, node := range nodes {
			node.Close()
		}
	}()
	assert.Equal(t, n, len(nodes))

	log.Println("after discovery")

	var pswg sync.WaitGroup
	data := []byte("data:my-topic")
	pswg.Add(1)
	sub1, err := Subscribe(nodes[0], "my-topic")
	assert.Nil(t, err)
	go func() {
		for {
			msg, err := sub1.Next(nodes[0].ctx)
			assert.Nil(t, err)
			assert.True(t, bytes.Equal(data, msg.Data))
			pswg.Done()
		}
	}()

	sub2, err := Subscribe(nodes[1],"other-topic")
	assert.Nil(t, err)
	go func() {
		for {
			sub2.Next(nodes[1].ctx)
			assert.Fail(t, "should not receive a message")
			return
		}
	}()

	topic3, err := Topic(nodes[2], "my-topic")
	assert.Nil(t, err)
	go func() {
		time.Sleep(1000 * time.Millisecond)
		topic3.Publish(nodes[2].ctx, data[:])
	}()
	pswg.Wait()
}

func createNode(psk pnet.PSK, onPeerFound OnPeerFound) *BaseNode {
	n := NewBaseNode(context.Background(), NewConfig(nil, psk), NewDiscoveryConfig(onPeerFound))
	log.Printf("new node: %s", n.Host().ID().Pretty())
	Connect(n, []peer.AddrInfo{}, true)
	return n
}

func setupNodesGroup(n int, psk pnet.PSK) ([]*BaseNode, error) {
	var discwg sync.WaitGroup
	discwg.Add(n)

	onPeerFound := OnPeerFoundWaitGroup(&discwg)
	nodes := []*BaseNode{}
	timeout := time.After(6 * time.Second)
	discovered := make(chan bool)

	i := n
	for i > 0 {
		i--
		node := createNode(psk, onPeerFound)
		if node == nil {
			return nil, errors.New("could not create node")
		}
		nodes = append(nodes, node)
	}

	go func() {
		discwg.Wait()
		discovered <- true
	}()

	select {
	case <-timeout:
		return nil, errors.New("setupNodesGroup timeout")
	case <-discovered:
		{
			actualPeers := nodes[n-1].Host().Peerstore().Peers()
			if len(actualPeers) != n {
				return nil, errors.New("could not connect to all peers")
			}
		}
	}
	return nodes, nil
}
