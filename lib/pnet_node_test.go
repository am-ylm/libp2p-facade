package lib

import (
	"context"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"log"
	"sync"
	"testing"
	"time"
)

func TestNewPrivateNetNode(t *testing.T) {
	psk := PNetSecret()
	n1, err := NewPrivateNetNode(NewOptions(nil, psk, nil))
	assert.Nil(t, err)
	err = n1.Dht.Bootstrap(context.Background())
	assert.Nil(t, err)

	priv2, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	n2, err := NewPrivateNetNode(NewOptions(priv2, psk, nil))
	assert.Nil(t, err)
	n1Info := peer.AddrInfo{
		ID:    n1.Node.ID(),
		Addrs: n1.Node.Addrs(),
	}
	conns := n2.ConnectToPeers([]peer.AddrInfo{n1Info}, true)
	for conn := range conns {
		assert.Nil(t, conn.Error)
	}
}

func TestDiscovery(t *testing.T) {
	var discwg sync.WaitGroup
	timeout := time.After(4 * time.Second)
	done := make(chan bool)
	// 3 nodes -> at least 3 discovery events
	n := 3
	discwg.Add(n)
	onPeerFound := func(pi peer.AddrInfo) bool {
		go func() {
			defer func() {
				// recover from calling Done on a negative wait group counter
				// this originates in a different behavior of discovery notifications cross OS
				if r := recover(); r != nil {
					return
				}
			}()
			discwg.Done()
		}()
		return true
	}
	psk := PNetSecret()
	n1, _ := NewPrivateNetNode(NewOptions(nil, psk, NewDiscoveryOptions(onPeerFound)))
	log.Printf("n1: %s", n1.Node.ID().Pretty())
	n1.ConnectToPeers([]peer.AddrInfo{}, true)

	n2, _ := NewPrivateNetNode(NewOptions(nil, psk, NewDiscoveryOptions(onPeerFound)))
	log.Printf("n2: %s", n2.Node.ID().Pretty())
	n2.ConnectToPeers([]peer.AddrInfo{}, true)

	time.Sleep(time.Duration(1000) * time.Millisecond)

	n3, _ := NewPrivateNetNode(NewOptions(nil, psk, NewDiscoveryOptions(onPeerFound)))
	log.Printf("n3: %s", n3.Node.ID().Pretty())
	n3.ConnectToPeers([]peer.AddrInfo{}, true)

	time.Sleep(time.Duration(1000) * time.Millisecond)

	go func() {
		discwg.Wait()
		peers3 := n3.Node.Peerstore().Peers()
		assert.Equal(t, n, len(peers3), "node should have peers of length n")
		done <- true
	}()

	select {
	case <-timeout:
		assert.Fail(t, "didn't receive enough discovery events")
	case <-done:
	}
}
