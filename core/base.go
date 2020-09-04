package core

import (
	"context"
	"errors"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/pnet"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"log"
)

type BaseNode struct {
	ctx context.Context

	priv crypto.PrivKey
	psk  pnet.PSK

	host host.Host

	dht   *kaddht.IpfsDHT
	store datastore.Batching

	ps *pubsub.PubSub
	topics map[string]*pubsub.Topic

	logger logging.EventLogger
}

func NewBaseNode(ctx context.Context, cfg *Config, discovery *DiscoveryConfig, opts ...libp2p.Option) *BaseNode {
	h, idht, err := NewBaseLibP2P(ctx, cfg, opts...)
	if err != nil {
		log.Panic("could not setup node")
	}
	var ps *pubsub.PubSub
	if discovery != nil {
		err = ConfigureDiscovery(ctx, h, discovery)
		if err == nil {
			ps, err = pubsub.NewGossipSub(ctx, h)
		}
		if err != nil {
			log.Panic("could not setup discovery / pubsub")
		}
	}
	n := BaseNode{
		ctx, cfg.PrivKey, cfg.Secret,
		h, idht, cfg.DS, ps,
		map[string]*pubsub.Topic{},
		cfg.Logger,
	}
	return &n
}

func (n *BaseNode) Context() context.Context {
	return n.ctx
}

func (n *BaseNode) Close() error {
	errs := Close(n)
	if len(errs) > 0 {
		return errors.New("could not close node")
	}
	return nil
}

func (n *BaseNode) PrivKey() crypto.PrivKey {
	return n.priv
}

func (n *BaseNode) Psk() pnet.PSK {
	return n.psk
}

func (n *BaseNode) Host() host.Host {
	return n.host
}

func (n *BaseNode) DHT() *kaddht.IpfsDHT {
	return n.dht
}

func (n *BaseNode) Store() datastore.Batching {
	return n.store
}

func (n *BaseNode) Logger() logging.EventLogger {
	return n.logger
}

func (n *BaseNode) PubSub() *pubsub.PubSub {
	return n.ps
}

func (n *BaseNode) Topics() map[string]*pubsub.Topic {
	return n.topics
}
