package core

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"sync"
	"time"
)

// ConnManagerConfig used to configure a connection manager (github.com/libp2p/go-libp2p-connmgr)
type ConnManagerConfig struct {
	Low   int
	High  int
	Grace time.Duration
}

// ConnectionResult is the used to abstract connection try
type PeerConnection struct {
	Error error
	Info  peer.AddrInfo
	Time  int64
}

// Connect will try to connect to each of the given peer
// a channel is used to publish connection result.
// the dht should be bootstrapped if the is the first connect of a node
func Connect(node LibP2PPeer, peers []peer.AddrInfo, bootDht bool) chan PeerConnection {
	connChannel := make(chan PeerConnection)
	var wg sync.WaitGroup
	for _, pinfo := range peers {
		wg.Add(1)
		go func(pinfo peer.AddrInfo) {
			defer wg.Done()
			err := node.Host().Connect(node.Context(), pinfo)
			if err != nil {
				node.Logger().Infof("new peer connected: %s", pinfo.ID.Pretty())
			}
			connChannel <- PeerConnection{err, pinfo, time.Now().Unix()}
		}(pinfo)
	}

	go func() {
		wg.Wait()
		if bootDht {
			node.DHT().Bootstrap(node.Context())
		}
		close(connChannel)
	}()

	return connChannel
}
