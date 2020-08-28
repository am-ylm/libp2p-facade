package lib

import (
	"context"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"

	"log"
	"sync"
	"testing"
)

// - create circuit-relay node
// - create node 1, connect to relayer
// - create node 2 + stream handler, connect to relayer
// - node 1 connect to node 2 via circuit-relay
func TestRelayer(t *testing.T) {
	var wg sync.WaitGroup

	psk := PNetSecret()

	priv, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	rel, _ := NewRelayer(context.Background(), NewOptions(priv, psk, nil))
	defer rel.Close()
	rel.Dht.Bootstrap(context.Background())
	relInfo := peer.AddrInfo{
		ID:    rel.Node.ID(),
		Addrs: rel.Node.Addrs(),
	}

	addr1, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/3031")
	n1, err := newNode(psk, []multiaddr.Multiaddr{addr1})
	defer n1.Close()
	assert.Nil(t, err)

	wg.Add(1)
	go func() {
		conns := n1.ConnectToPeers([]peer.AddrInfo{relInfo}, true)
		for conn := range conns {
			log.Println("new connection " + conn.ID.String() + ", error: ", conn.Error)
		}
		wg.Done()
	}()
	wg.Wait()

	addr2, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/3032")
	n2, err := newNode(psk, []multiaddr.Multiaddr{addr2})
	defer n1.Close()
	assert.Nil(t, err)
	n2.Node.SetStreamHandler("/hello", func(s network.Stream) {
		wg.Done()
		s.Close()
	})
	// n2 -> rel
	wg.Add(1)
	go func() {
		conns := n2.ConnectToPeers([]peer.AddrInfo{relInfo}, true)
		for conn := range conns {
			log.Println("new connection " + conn.ID.String() + ", error: ", conn.Error)
		}
		wg.Done()
	}()
	wg.Wait()

	// n1 -> relay -> n2
	n2relayInfo := CircuitRelayAddrInfo(rel.Node.ID(), n2.Node.ID())
	wg.Add(1)
	go func() {
		conns := n1.ConnectToPeers([]peer.AddrInfo{n2relayInfo}, false)
		for conn := range conns {
			log.Println("new connection " + conn.ID.String() + ", error: ", conn.Error)
		}
		wg.Done()
	}()
	wg.Wait()

	wg.Add(1)
	s, err := n1.Node.NewStream(context.Background(), n2.Node.ID(), "/hello")
	assert.Nil(t, err, "can't send message: %s", err)
	s.Read(make([]byte, 1)) // block until the handler closes the stream
	wg.Wait()
}

func newNode(psk pnet.PSK, addrs []multiaddr.Multiaddr) (*PrivateNetNode, error) {
	priv, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	ropts := NewOptions(priv, psk, nil)
	ropts.Addrs = addrs
	ropts.UseLibp2pOpts = func(_opts []libp2p.Option) ([]libp2p.Option, error) {
		return append(_opts,
			libp2p.EnableRelay(),
			libp2p.ConnectionManager(connmgr.NewConnManager(10, 50, ConnectionsGrace)),
		), nil
	}
	return NewPrivateNetNode(context.Background(), ropts)
}
