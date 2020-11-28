package core

import (
	"context"
	"errors"
	"sync"
	"time"

	logging "github.com/ipfs/go-log/v2"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DefaultDiscoveryInterval = time.Minute * 10

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DefaultDiscoveryServiceTag = "pnet:pubsub"

// OnPeerFound will be triggered on new peer discovery
// in case it returns false, this node won't connect to the given peer
type OnPeerFound = func(pi peer.AddrInfo) bool

// DiscoveryConfig
type DiscoveryConfig struct {
	OnPeerFound OnPeerFound
	ServiceTag  string
	Interval    time.Duration
	Services    []discovery.Service
}

// NewDiscoveryConfig creates a new discovery config object with defaults
func NewDiscoveryConfig(onPeerFound OnPeerFound) *DiscoveryConfig {
	if onPeerFound == nil {
		onPeerFound = func(pi peer.AddrInfo) bool {
			return true
		}
	}
	opts := DiscoveryConfig{
		onPeerFound,
		DefaultDiscoveryServiceTag,
		DefaultDiscoveryInterval,
		[]discovery.Service{},
	}
	return &opts
}

// SetupGroup will create a group of n local nodes that are connected to each other
// assuming (at least) n-1 will be connected otherwise an error is returned
// n-1 for stability as this function is used in tests
func SetupGroup(n int, nodeFactory NodeFactory) ([]LibP2PPeer, error) {
	var discwg sync.WaitGroup

	nodes := []LibP2PPeer{}
	peers := []peer.AddrInfo{}
	timeout := time.After(5 * time.Second)
	discovered := make(chan bool, 1)

	i := n
	for i > 0 {
		i--
		node := nodeFactory()
		if node == nil {
			return nil, errors.New("could not create node")
		}
		go AutoClose(node.Context(), node)
		discwg.Add(len(nodes))
		nodes = append(nodes, node)
		go func() {
			defer func() {
				// recover from calling Done on a negative wait group counter
				// this originates in a different behavior of discovery notifications cross OS
				if r := recover(); r != nil {
					return
				}
			}()
			conns := Connect(node, peers, true)
			for conn := range conns {
				node.Logger().Info("connect event:", conn)
				discwg.Done()
			}
		}()
		peers = append(peers, peer.AddrInfo{
			ID:    node.Host().ID(),
			Addrs: node.Host().Addrs(),
		})
	}

	go func() {
		select {
		case <-timeout:
			panic("setupNodesGroup timeout")
		case <-discovered:
			// do nothing
		}
	}()

	discwg.Wait()

	discovered <- true

	actualPeers := nodes[0].Host().Peerstore().Peers()
	if len(actualPeers) < n-1 {
		return nil, errors.New("could not connect to all peers")
	}

	return nodes, nil
}

// ConfigureDiscovery binds mDNS discovery services
func ConfigureDiscovery(ctx context.Context, h host.Host, opts *DiscoveryConfig, logger logging.EventLogger) error {
	discoveryServices := opts.Services

	// setup default mDNS discovery to find local peers
	disc, err := discovery.NewMdnsService(ctx, h, opts.Interval, opts.ServiceTag)
	// if couldn't setup local mDNS and no other service was provided -> exit
	if err != nil && len(discoveryServices) == 0 {
		return err
	}
	discoveryServices = append(discoveryServices, disc)

	n := discoveryNotifee{h, ctx, opts.OnPeerFound, logger}
	for _, disc := range discoveryServices {
		disc.RegisterNotifee(&n)
	}

	return nil
}

type discoveryNotifee struct {
	h           host.Host
	ctx         context.Context
	onPeerFound OnPeerFound
	logger      logging.EventLogger
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.logger.Infof("node %s", n.h.ID().Pretty())
	printPeer(n.logger, "discovered new peer", pi)
	if n.onPeerFound == nil || n.onPeerFound(pi) {
		err := n.h.Connect(n.ctx, pi)
		if err != nil {
			n.logger.Warnf("could not connect to peer %s: %s\n", pi.ID.Pretty(), err)
		} else {
			n.logger.Infof("connected to peer %s", pi.ID.Pretty())
		}
	}
}

func printPeer(logger logging.EventLogger, prefix string, pi peer.AddrInfo) {
	id := pi.ID.Pretty()
	logger.Infof("%s %s, listening on:", prefix, id)
	for _, addr := range pi.Addrs {
		logger.Infof("\t- %s", addr.String())
	}
}
