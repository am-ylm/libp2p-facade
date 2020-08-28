package lib

import (
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

func TestDiscovery(t *testing.T) {
	n := 3
	psk := PNetSecret()
	nodes, err := setupNodesGroup(n, psk)
	assert.Nil(t, err)
	assert.Equal(t, n, len(nodes))
}

func createNode(psk pnet.PSK, onPeerFound OnPeerFound) *PrivateNetNode {
	n, err := NewPrivateNetNode(context.Background(), NewOptions(nil, psk, NewDiscoveryOptions(onPeerFound)))
	if err != nil {
		log.Fatalf("could not create node: %s", err.Error())
		return nil
	}
	log.Printf("new node: %s", n.Node.ID().Pretty())
	n.ConnectToPeers([]peer.AddrInfo{}, true)
	return n
}

func setupNodesGroup(n int, psk pnet.PSK) ([]*PrivateNetNode, error) {
	var discwg sync.WaitGroup
	discwg.Add(n)

	onPeerFound := OnPeerFoundWaitGroup(&discwg)
	nodes := []*PrivateNetNode{}
	timeout := time.After(5 * time.Second)
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
	case <-discovered: {
		actualPeers := nodes[n-1].Node.Peerstore().Peers()
		if len(actualPeers) != n {
			return nil, errors.New("could not connect to all peers")
		}
	}
	}
	return nodes, nil
}