package p2pfacade

import (
	"context"
	"sync"
	"time"

	logging "github.com/ipfs/go-log/v2"
	libp2pnetwork "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	// notiffeeCacheGCInterval determines how frequent the GC is working
	notiffeeCacheGCInterval = time.Minute * 15
	// connectTimeout is the timeout used for connections
	connectTimeout = time.Minute
	// backoffLow is when we start the backoff exponent interval
	backoffLow = 10 * time.Second
	// backoffLow is when we stop the backoff exponent interval
	backoffHigh = 30 * time.Minute
	// backoffExponentBase is the base of the backoff exponent
	backoffExponentBase = 2.0
	// backoffConnectorCacheSize is the cache size of the backoff connector
	backoffConnectorCacheSize = 1024
	// connectorQueueSize is the buffer size of the channel used by the connector
	connectorQueueSize = 32
)

var (
	loggerConn = logging.Logger("p2p:conn")
)

// startConnector starts to receive and handle incoming connect requests
func (f *facade) startConnector(connectQ ConnectQueue) {
	buffer := make(chan peer.AddrInfo, connectorQueueSize)

	go func() {
		ctx, cancel := context.WithCancel(f.ctx)
		defer cancel()
		f.backoffConnector.Connect(ctx, buffer)
	}()

	go func() {
		ctx, cancel := context.WithCancel(f.ctx)
		defer cancel()
		for {
			select {
			case pi := <-connectQ:
				switch f.host.Network().Connectedness(pi.ID) {
				case libp2pnetwork.CannotConnect, libp2pnetwork.Connected:
					return
				default:
				}
				loggerConn.Debugf("found new peer %s", pi.String())
				select {
				case buffer <- pi:
				default:
					// TODO: handle skipped peer
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func Notiffee(net libp2pnetwork.Network) (*libp2pnetwork.NotifyBundle, func()) {
	connectedCache := map[peer.ID]bool{}
	l := &sync.RWMutex{}

	gc := func() {
		l.Lock()
		defer l.Unlock()

		toRemove := make([]peer.ID, 0)
		for pid := range connectedCache {
			switch net.Connectedness(pid) {
			case libp2pnetwork.CannotConnect, libp2pnetwork.NotConnected:
				toRemove = append(toRemove, pid)
			default:
			}
		}
		for _, pid := range toRemove {
			delete(connectedCache, pid)
		}
	}

	return &libp2pnetwork.NotifyBundle{
		ConnectedF: func(n libp2pnetwork.Network, c libp2pnetwork.Conn) {
			l.Lock()
			defer l.Unlock()

			pid := c.RemotePeer()
			if _, ok := connectedCache[pid]; !ok {
				connectedCache[pid] = true
				metricConnections.WithLabelValues(n.LocalPeer().String()).Inc()
				loggerConn.Debugf("new connected peer %s", pid.String())
			}
		},
		DisconnectedF: func(n libp2pnetwork.Network, c libp2pnetwork.Conn) {
			l.Lock()
			defer l.Unlock()

			pid := c.RemotePeer()
			if n.Connectedness(pid) == libp2pnetwork.Connected {
				return
			}
			if _, ok := connectedCache[pid]; !ok {
				delete(connectedCache, pid)
				metricConnections.WithLabelValues(n.LocalPeer().String()).Dec()
				loggerConn.Debugf("disconnected peer %s", pid.String())
			}
		},
	}, gc
}
