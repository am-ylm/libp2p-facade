package core

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

var defaultRelayConnectionConfig = ConnManagerConfig{
	200, 1000, time.Minute,
}

// NewRelayer creates an instance of NewBaseNode with the needed configuration to be a circuit-relay node
func NewRelayer(ctx context.Context, cfg *Config, disc *DiscoveryConfig, opts ...libp2p.Option) LibP2PNode {
	if cfg.ConnManagerConfig == nil {
		cfg.ConnManagerConfig = &defaultRelayConnectionConfig
	}
	opts = append(opts, libp2p.EnableRelay(circuit.OptHop))
	return NewBaseNode(ctx, cfg, disc, opts...)
}

// CircuitRelayAddr construct a circuit relay address of the given relay and target peer
func CircuitRelayAddr(relay, target peer.ID) multiaddr.Multiaddr {
	rawAddr := fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s", relay.Pretty(), target.Pretty())
	addr, _ := multiaddr.NewMultiaddr(rawAddr)
	return addr
}

func CircuitRelayAddrInfo(relay, target peer.ID) peer.AddrInfo {
	return peer.AddrInfo{
		ID: target,
		Addrs: []multiaddr.Multiaddr{CircuitRelayAddr(relay, target)},
	}
}
