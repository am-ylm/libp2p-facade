package main

import (
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"

	pnet_node "github.com/amirylm/go-libp2p-pnet-node"
)

var (
	ConnectionsLow   = 200
	ConnectionsHigh  = 1000
	ConnectionsGrace = time.Minute
)

// NewRelayer creates an instance of PrivateNetNode with the needed configuration to be a circuit-relay node
func NewRelayer(opts *pnet_node.Options) (*pnet_node.PrivateNetNode, error) {
	orig := opts.Libp2pOpts
	opts.Libp2pOpts = func() ([]libp2p.Option, error) {
		o := []libp2p.Option{}
		var err error
		if orig != nil {
			o, err = orig()
		}
		o = append(o,
			libp2p.ConnectionManager(connmgr.NewConnManager(ConnectionsLow, ConnectionsHigh, ConnectionsGrace)),
			libp2p.EnableRelay(circuit.OptHop),
		)
		return o, err
	}
	return pnet_node.NewPrivateNetNode(opts)
}

func CircuitRelayAddr(relay, target peer.ID) multiaddr.Multiaddr {
	rawAddr := fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s", relay.Pretty(), target.Pretty())
	addr, _ := multiaddr.NewMultiaddr(rawAddr)
	return addr
}

func CircuitRelayAddrInfo(relay, target peer.ID) peer.AddrInfo {
	return peer.AddrInfo{
		ID:    target,
		Addrs: []multiaddr.Multiaddr{CircuitRelayAddr(relay, target)},
	}
}
