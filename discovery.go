package p2pfacade

import (
	"context"

	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	libp2pdisc "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/pkg/errors"
)

// NewKadDHT creates a new kademlia DHT and a corresponding discovery service
func NewKadDHT(ctx context.Context, host host.Host, protocolPrefix protocol.ID,
	mode dht.ModeOpt, bootstrappers []peer.AddrInfo) (routing.Routing, discovery.Discovery, error) {
	kdht, err := dht.New(ctx, host, dht.ProtocolPrefix(protocolPrefix), dht.Mode(mode), dht.BootstrapPeers(bootstrappers...))
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not create DHT")
	}
	logger.Debugf("created Kademlia DHT with protocol prefix %s", string(protocolPrefix))
	return kdht, libp2pdisc.NewRoutingDiscovery(kdht), nil
}

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
