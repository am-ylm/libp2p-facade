package p2pfacade

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// New creates a new mdns service
func NewMdns(ctx context.Context, connect ConnectQueue, host host.Host, serviceTag string) mdns.Service {
	md := mdnsDisc{ctx, connect}

	return mdns.NewMdnsService(host, serviceTag, &md)
}

type mdnsDisc struct {
	ctx     context.Context
	connect ConnectQueue
}

// HandlePeerFound implements mdns.Notifee
func (md *mdnsDisc) HandlePeerFound(pi peer.AddrInfo) {
	if md.ctx.Err() == nil {
		select {
		case md.connect <- pi:
		default:
		}
	}
}
