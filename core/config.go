package core

import (
	"context"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-ipns"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	record "github.com/libp2p/go-libp2p-record"
	secio "github.com/libp2p/go-libp2p-secio"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/multiformats/go-multiaddr"
	"time"
)

// BootstrapLibP2P creates an instance of libp2p host + DHT and peer discovery / pubsub (if configured)
func BootstrapLibP2P(ctx context.Context, cfg *Config, opts ...libp2p.Option) (host.Host, *kaddht.IpfsDHT, *pubsub.PubSub, error) {
	var idht *kaddht.IpfsDHT
	var err error

	libp2pOpts, err := cfg.ToLibP2pOpts(opts...)
	if err != nil {
		return nil, nil, nil, err
	}

	libp2pOpts = append(libp2pOpts,
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = newDHT(ctx, h, cfg.DS)
			return idht, err
		}),
	)

	h, err := libp2p.New(
		ctx,
		libp2pOpts...,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	err = ConfigureDiscovery(ctx, h, cfg.Discovery, cfg.Logger)
	if err != nil {
		return h, idht, nil, err
	}
	ps, err := pubsub.NewGossipSub(ctx, h)

	return h, idht, ps, err
}

// Config holds the needed configuration for creating a private node instance
type Config struct {
	// PrivKey of the current node
	PrivKey crypto.PrivKey
	// Secret is the private network secret ([32]byte)
	Secret pnet.PSK
	// Addrs are Multiaddrs for the current node, will fallback to libp2p defaults
	Addrs []multiaddr.Multiaddr
	// Logger to use (see github.com/ipfs/go-log/v2)
	Logger logging.EventLogger
	// DS is the main data store used by DHT, BlockStore, etc...
	DS datastore.Batching
	// Peers are nodes that we want to connect on bootstrap
	Peers []peer.AddrInfo
	// ConnManagerConfig is used to configure the conn management of current peer
	ConnManagerConfig *ConnManagerConfig
	// Discovery is used to configure discovery (+pubsub)
	Discovery *DiscoveryConfig
}

// NewConfig creates the minimum needed Config
func NewConfig(priv crypto.PrivKey, psk pnet.PSK, store datastore.Batching) *Config {
	opts := Config{
		PrivKey: priv,
		Secret:  psk,
		DS: store,
	}
	return &opts
}

// ToLibP2pOpts converts Config into the corresponding []libp2p.Option
func (cfg *Config) ToLibP2pOpts(customOpts ...libp2p.Option) ([]libp2p.Option, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, err
	}
	libp2pOpts := []libp2p.Option{
		libp2p.Identity(cfg.PrivKey),
		libp2p.ListenAddrs(cfg.Addrs...),
		libp2p.PrivateNetwork(cfg.Secret),
		libp2p.NATPortMap(),
		libp2p.DefaultEnableRelay,
		libp2p.ConnectionManager(connmgr.NewConnManager(cfg.ConnManagerConfig.Low, cfg.ConnManagerConfig.High, cfg.ConnManagerConfig.Grace)),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(secio.ID, secio.New),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
	}
	return append(libp2pOpts, customOpts...), nil
}

func (cfg *Config) defaults() error {
	if cfg.PrivKey == nil {
		priv, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
		cfg.PrivKey = priv
	}
	if cfg.Secret == nil {
		cfg.Secret = PNetSecret()
	}
	if cfg.DS == nil {
		cfg.DS = dssync.MutexWrap(datastore.NewMapDatastore())
	}
	if cfg.Addrs == nil || len(cfg.Addrs) == 0 {
		addripv4, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/0")
		addripv6, _ := multiaddr.NewMultiaddr("/ip6/::/tcp/0")
		cfg.Addrs = []multiaddr.Multiaddr{addripv4, addripv6}
	}
	if cfg.ConnManagerConfig == nil {
		cmc := ConnManagerConfig{100, 600, time.Minute}
		cfg.ConnManagerConfig = &cmc
	}
	if cfg.Discovery == nil {
		cfg.Discovery = NewDiscoveryConfig(nil)
	}
	if cfg.Logger == nil {
		cfg.Logger = defaultLogger()
	}
	return nil
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