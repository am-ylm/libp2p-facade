package lib

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DefaultDiscoveryInterval = time.Minute * 10

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DefaultDiscoveryServiceTag = "pnet:cpubsub"

// OnPeerFound will be triggered on new peer discovery
// in case it returns false, this node won't connect to the given peer
type OnPeerFound = func(pi peer.AddrInfo) bool

// DiscoveryOptions
type DiscoveryOptions struct {
	OnPeerFound OnPeerFound
	ServiceTag  string
	Interval    time.Duration
	Services    []discovery.Service
	Host        host.Host
	Ctx         context.Context
}

// NewDiscoveryOptions creates a default options object
func NewDiscoveryOptions(onPeerFound OnPeerFound) *DiscoveryOptions {
	if onPeerFound == nil {
		onPeerFound = func(pi peer.AddrInfo) bool {
			return true
		}
	}
	opts := DiscoveryOptions{
		onPeerFound,
		DefaultDiscoveryServiceTag,
		DefaultDiscoveryInterval,
		[]discovery.Service{},
		nil,
		context.Background(),
	}
	return &opts
}

// configureDiscovery binds mDNS discovery services
func configureDiscovery(opts *DiscoveryOptions) error {
	discoveryServices := opts.Services

	// setup default mDNS discovery to find local peers
	disc, err := discovery.NewMdnsService(context.Background(), opts.Host, opts.Interval, opts.ServiceTag)
	// if couldn't setup local mDNS and no other service was provided -> exit
	if err != nil && len(discoveryServices) == 0 {
		return err
	}
	discoveryServices = append(discoveryServices, disc)

	for _, disc := range discoveryServices {
		n := discoveryNotifee{opts.Host, opts.Ctx, opts.OnPeerFound}
		disc.RegisterNotifee(&n)
	}

	return nil
}

type discoveryNotifee struct {
	h           host.Host
	ctx         context.Context
	onPeerFound OnPeerFound
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	log.Printf("node %s", n.h.ID().Pretty())
	printPeer("discovered new peer", pi)
	if n.onPeerFound == nil || n.onPeerFound(pi) {
		err := n.h.Connect(n.ctx, pi)
		if err != nil {
			log.Printf("could not connect to peer %s: %s\n", pi.ID.Pretty(), err)
		} else {
			log.Printf("connected to peer %s", pi.ID.Pretty())
		}
	}
}

func printPeer(prefix string, pi peer.AddrInfo) {
	id := pi.ID.Pretty()
	log.Printf("%s %s, listening on:", prefix, id)
	for _, addr := range pi.Addrs {
		log.Printf("\t- %s", addr.String())
	}
}
