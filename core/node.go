package core

import (
	"context"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/pnet"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
)

// LibP2PNode is the base interface
type LibP2PNode interface {
	Context() context.Context

	Close() error

	PrivKey() crypto.PrivKey
	Psk() pnet.PSK

	Host() host.Host
	DHT() *kaddht.IpfsDHT
	Store() datastore.Batching

	PubSuber

	Logger() logging.EventLogger
}

func Close(node LibP2PNode) []error {
	all := []error{
		node.DHT().Close(),
		node.Host().Close(),
	}

	errs := []error{}
	for _, err := range all {
		if err != nil {
			node.Logger().Warnf("could not close node: %s", err.Error())
			errs = append(errs, err)
		}
	}

	return errs
}

