package p2pfacade

import (
	"context"
	"io"
	"math/rand"
	"time"

	"github.com/amirylm/libp2p-facade/commons"
	"github.com/amirylm/libp2p-facade/pubsub"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	libp2pdisc "github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type ConnectQueue chan peer.AddrInfo

// Facade is an interface on top of libp2p
type Facade interface {
	Start(connectQ ConnectQueue) error
	Host() host.Host
	pubsub.PubsubService
	io.Closer
}

// StartNodes spins up nodes according to given config
func StartNodes(ctx context.Context, cfgs []*commons.Config) ([]Facade, error) {
	nodes := []Facade{}

	for _, cfg := range cfgs {
		f, err := New(ctx, cfg)
		if err != nil {
			return nodes, err
		}
		nodes = append(nodes, f)

		err = f.Start(nil)
		if err != nil {
			return nodes, err
		}
	}

	return nodes, nil
}

// New creates a new p2p facade with the given config,
// if options were provided they will be used instead.
func New(ctx context.Context, cfg *commons.Config, opts ...libp2p.Option) (Facade, error) {
	if len(opts) == 0 {
		err := cfg.Init()
		if err != nil {
			return nil, err
		}
		opts, err = cfg.Libp2pOptions()
		if err != nil {
			return nil, err
		}
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}
	f := facade{
		ctx:  ctx,
		host: h,
		cfg:  cfg,
	}

	n, gc := Notiffee(h.Network())

	h.Network().Notify(n)
	go func() {
		for ctx.Err() == nil {
			time.Sleep(notiffeeCacheGCInterval)
			gc()
		}
	}()

	backoffFactory := libp2pdisc.NewExponentialDecorrelatedJitter(
		backoffLow, backoffHigh, backoffExponentBase, rand.NewSource(0))
	backoffConnector, err := libp2pdisc.NewBackoffConnector(f.host, backoffConnectorCacheSize, connectTimeout, backoffFactory)
	if err != nil {
		return &f, err
	}
	f.backoffConnector = backoffConnector

	if len(f.cfg.MdnsServiceTag) > 0 {
		f.mdnsq = make(ConnectQueue)
		f.mdnsSvc = NewMdns(ctx, f.mdnsq, f.host, f.cfg.MdnsServiceTag)
	}

	if cfg.Pubsub != nil {
		err = f.setupPubsub()
	}

	return &f, err
}

type facade struct {
	ctx  context.Context
	cfg  *commons.Config
	host host.Host
	ps   pubsub.PubsubService

	routing          routing.Routing
	backoffConnector *libp2pdisc.BackoffConnector

	mdnsSvc mdns.Service
	mdnsq   ConnectQueue
	// relayers []peer.AddrInfo
}

func (f *facade) Start(connectQ ConnectQueue) error {
	if f.mdnsSvc != nil {
		if err := f.mdnsSvc.Start(); err != nil {
			return err
		}
		f.startConnector(f.mdnsq)
	}

	if connectQ != nil {
		f.startConnector(connectQ)
	}

	if f.routing != nil {
		if err := f.routing.Bootstrap(f.ctx); err != nil {
			return err
		}
	}

	return nil
}

func (f *facade) Host() host.Host {
	return f.host
}

func (f *facade) Close() error {
	if err := f.host.Close(); err != nil {
		return err
	}
	return nil
}
