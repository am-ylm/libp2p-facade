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

func Connect(node LibP2PNode, peers []peer.AddrInfo, bootDht bool) chan PeerConnection {
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