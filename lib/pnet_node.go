package lib

import (
	"context"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipns"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	record "github.com/libp2p/go-libp2p-record"
	"sync"
	"time"
)

var (
	ConnectionsLow   = 100
	ConnectionsHigh  = 500
	ConnectionsGrace = time.Minute
)

// ConnectionResult is the used to abstract connection try
type ConnectionResult struct {
	Error error
	ID    peer.ID
}

// PrivateNetNode holds a libp2p node and dht instaces
// it abstract the needed configuration and setup for libp2p
type PrivateNetNode struct {
	ctx  context.Context
	Node host.Host

	Dht *kaddht.IpfsDHT

	logger	 logging.EventLogger
}

// NewPrivateNetNode creates an instance of PrivateNetNode
func NewPrivateNetNode(opts *Options) (*PrivateNetNode, error) {
	h, dht, err := SetupLibp2p(opts)
	if err != nil {
		return nil, err
	}
	n := PrivateNetNode{opts.Ctx, h, dht, opts.Logger}
	return &n, nil
}

// Close closes the involved components
func (rel *PrivateNetNode) Close() []error {
	return []error{
		rel.Dht.Close(),
		rel.Node.Close(),
	}
}

// ConnectToPeers will try to connect to all given peers.
// the results are channeled and should be handled in the caller.
// DHT will be bootstrapped as well
func (n *PrivateNetNode) ConnectToPeers(peers []peer.AddrInfo, bootDht bool) chan ConnectionResult {
	connChannel := make(chan ConnectionResult)
	var wg sync.WaitGroup
	for _, pinfo := range peers {
		wg.Add(1)
		go func(pinfo peer.AddrInfo) {
			defer wg.Done()
			err := n.Node.Connect(n.ctx, pinfo)
			if err != nil {
				n.logger.Infof("new peer connected: %s", pinfo.ID.Pretty())
			}
			connChannel <- ConnectionResult{err, pinfo.ID}
		}(pinfo)
	}

	go func() {
		wg.Wait()
		if bootDht {
			n.Dht.Bootstrap(n.ctx)
		}
		close(connChannel)
	}()

	return connChannel
}

// SetupLibp2p will configure all the liibp2p related stuff
func SetupLibp2p(opts *Options) (host.Host, *kaddht.IpfsDHT, error) {
	var idht *kaddht.IpfsDHT
	var err error

	libp2pOpts, err := opts.ToLibP2pOpts()
	if err != nil {
		return nil, nil, err
	}

	libp2pOpts = append(libp2pOpts,
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = newDHT(opts.Ctx, h, opts.DS)
			return idht, err
		}),
	)

	h, err := libp2p.New(
		opts.Ctx,
		libp2pOpts...,
	)
	if err != nil {
		return nil, nil, err
	}

	if opts.Discovery != nil {
		opts.Discovery.Host = h
		err = configureDiscovery(opts.Discovery)
	}

	return h, idht, err
}

func newDHT(ctx context.Context, h host.Host, ds datastore.Batching) (*kaddht.IpfsDHT, error) {
	dhtOpts := []kaddht.Option{
		kaddht.NamespacedValidator("pk", record.PublicKeyValidator{}),
		kaddht.NamespacedValidator("ipns", ipns.Validator{KeyBook: h.Peerstore()}),
		kaddht.Concurrency(10),
		kaddht.Mode(kaddht.ModeAuto),
	}
	if ds != nil {
		dhtOpts = append(dhtOpts, kaddht.Datastore(ds))
	}

	return kaddht.New(ctx, h, dhtOpts...)
}