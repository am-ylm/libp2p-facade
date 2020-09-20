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

// LibP2PPeer is the base interface
type LibP2PPeer interface {
	Context() context.Context

	Closeable

	PrivKey() crypto.PrivKey
	Psk() pnet.PSK

	Host() host.Host
	DHT() *kaddht.IpfsDHT
	Store() datastore.Batching

	PubSuber

	Logger() logging.EventLogger
}

// Closeable represent an object the can be closed
type Closeable interface {
	Close() error
}

// AutoClose closes the given Closeable once the context is done
func AutoClose(ctx context.Context, c Closeable) {
	select {
	case <-ctx.Done():
		c.Close()
	}
}

// Close closes the underlying host, dht and possibly other services (e.g. store, pubsub)
// returns an array to be compatible with multiple closes
func Close(node LibP2PPeer) []error {
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
