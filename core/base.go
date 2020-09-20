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

type BasePeer struct {
	ctx context.Context

	priv crypto.PrivKey
	psk  pnet.PSK

	host host.Host

	dht   *kaddht.IpfsDHT
	store datastore.Batching

	ps     *pubsub.PubSub
	topics map[string]*pubsub.Topic

	logger logging.EventLogger
}

func NewBasePeer(ctx context.Context, cfg *Config, opts ...libp2p.Option) *BasePeer {
	h, idht, ps, err := BootstrapLibP2P(ctx, cfg, opts...)
	if err != nil {
		log.Panic("could not setup peer")
	}
	p := BasePeer{
		ctx, cfg.PrivKey, cfg.Secret,
		h, idht, cfg.DS, ps,
		map[string]*pubsub.Topic{},
		cfg.Logger,
	}
	return &p
}

func (p *BasePeer) Context() context.Context {
	return p.ctx
}

func (p *BasePeer) Close() error {
	errs := Close(p)
	if len(errs) > 0 {
		return errors.New("could not close node")
	}
	return nil
}

func (p *BasePeer) PrivKey() crypto.PrivKey {
	return p.priv
}

func (p *BasePeer) Psk() pnet.PSK {
	return p.psk
}

func (p *BasePeer) Host() host.Host {
	return p.host
}

func (p *BasePeer) DHT() *kaddht.IpfsDHT {
	return p.dht
}

func (p *BasePeer) Store() datastore.Batching {
	return p.store
}

func (p *BasePeer) Logger() logging.EventLogger {
	return p.logger
}

func (p *BasePeer) PubSub() *pubsub.PubSub {
	return p.ps
}

func (p *BasePeer) Topics() map[string]*pubsub.Topic {
	return p.topics
}
