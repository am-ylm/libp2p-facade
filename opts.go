package go_libp2p_pnet_node

import (
	"context"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/pnet"
	secio "github.com/libp2p/go-libp2p-secio"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/multiformats/go-multiaddr"
	"math/rand"
)

type ConfigureLibp2pOpts = func() ([]libp2p.Option, error)

// Options holds the needed configuration for creating a private node instance
type Options struct {
	Ctx      context.Context
	PrivKey  crypto.PrivKey
	Secret   pnet.PSK
	Addrs    []multiaddr.Multiaddr
	Logger	 logging.EventLogger
	DS       datastore.Batching
	Libp2pOpts ConfigureLibp2pOpts
}

// NewOptions creates the minimum needed Options
func NewOptions(priv crypto.PrivKey, psk pnet.PSK) *Options {
	opts := Options{
		Ctx:     context.Background(),
		PrivKey: priv,
		Secret:  psk,
	}
	return &opts
}

// ToLibP2pOpts converts Options into the corresponding []libp2p.Option
func (opts *Options) ToLibP2pOpts() ([]libp2p.Option, error) {
	err := opts.defaults()
	if err != nil {
		return nil, err
	}
	customOpts, err := opts.Libp2pOpts()
	if err != nil {
		return nil, err
	}
	return append([]libp2p.Option{
		libp2p.Identity(opts.PrivKey),
		libp2p.ListenAddrs(opts.Addrs...),
		libp2p.PrivateNetwork(opts.Secret),
		libp2p.NATPortMap(),
		libp2p.EnableAutoRelay(),
		libp2p.EnableNATService(),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(secio.ID, secio.New),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
	}, customOpts...), nil
}

func (opts *Options) defaults() error {
	if opts.PrivKey == nil {
		priv, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
		opts.PrivKey = priv
	}
	if opts.Secret == nil {
		opts.Secret = RandSecret()
	}
	if opts.Addrs == nil || len(opts.Addrs) == 0 {
		// currently using libp2p defaults, might be changed
		addripv4, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/0")
		addripv6, _ := multiaddr.NewMultiaddr("/ip6/::/tcp/0")
		opts.Addrs = []multiaddr.Multiaddr{addripv4, addripv6}
	}
	if opts.Logger == nil {
		opts.Logger = defaultLogger()
	}
	if opts.Libp2pOpts == nil {
		opts.Libp2pOpts = func() ([]libp2p.Option, error) {
			return []libp2p.Option{}, nil
		}
	}
	return nil
}

// RandSecret creates a new random secret, 32 length
func RandSecret() []byte {
	return randBytes(32)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
// randBytes creates a random string in the given length
func randBytes(n int) []byte {
	b := make([]byte, n)
	letters := len(letterBytes)
	for i := range b {
		b[i] = letterBytes[rand.Intn(letters)]
	}
	return b[:]
}

func defaultLogger() logging.EventLogger {
	return logging.Logger("libp2p-pnet-node")
}