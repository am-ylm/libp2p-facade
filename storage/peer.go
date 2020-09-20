package storage

import (
	"context"
	"errors"
	"github.com/amirylm/priv-libp2p-node/core"
	"github.com/ipfs/go-bitswap"
	"github.com/ipfs/go-bitswap/network"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	exoffline "github.com/ipfs/go-ipfs-exchange-offline"
	provider "github.com/ipfs/go-ipfs-provider"
	"github.com/ipfs/go-ipfs-provider/queue"
	"github.com/ipfs/go-ipfs-provider/simple"
	ipld "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipfs/go-merkledag"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/pnet"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"time"
)

const (
	defaultReprovideInterval = 8 * time.Hour
)

//
type StoragePeer interface {
	core.LibP2PPeer

	DagService() ipld.DAGService
	BlockService() blockservice.BlockService
	Reprovider() provider.System
}

type storagePeer struct {
	ipld.DAGService

	base *core.BasePeer

	bsrv       blockservice.BlockService
	reprovider provider.System
}

func NewStoragePeer(base *core.BasePeer, offline bool) StoragePeer {
	if offline {
		dag, bsrv, _, _ := CreateOfflineDagServices(base)
		repro := provider.NewOfflineProvider()
		node := storagePeer{dag, base, bsrv, repro}
		return &node
	}
	dag, bsrv, bstore, err := CreateDagServices(base)
	if err != nil {
		base.Logger().Panic("could not create DAG services")
		return nil
	}
	repro, err := SetupReprovider(base, bstore, defaultReprovideInterval)
	node := storagePeer{dag, base, bsrv, repro}
	return &node
}

func CreateDagServices(base core.LibP2PPeer) (ipld.DAGService, blockservice.BlockService, blockstore.Blockstore, error) {
	var bsrv blockservice.BlockService
	var bs blockstore.Blockstore

	bs = blockstore.NewBlockstore(base.Store())
	bs = blockstore.NewIdStore(bs)
	bs, _ = blockstore.CachedBlockstore(base.Context(), bs, blockstore.DefaultCacheOpts())

	bswapnet := network.NewFromIpfsHost(base.Host(), base.DHT())
	bswap := bitswap.New(base.Context(), bswapnet, bs)
	bsrv = blockservice.New(bs, bswap)

	return merkledag.NewDAGService(bsrv), bsrv, bs, nil
}

func CreateOfflineDagServices(base core.LibP2PPeer) (ipld.DAGService, blockservice.BlockService, blockstore.Blockstore, error) {
	bs := blockstore.NewBlockstore(base.Store())
	bs = blockstore.NewIdStore(bs)
	bs, _ = blockstore.CachedBlockstore(base.Context(), bs, blockstore.DefaultCacheOpts())

	bsrv := blockservice.New(bs, exoffline.Exchange(bs))

	return merkledag.NewDAGService(bsrv), bsrv, bs, nil
}

func SetupReprovider(base core.LibP2PPeer, bstore blockstore.Blockstore, reprovideInterval time.Duration) (provider.System, error) {
	queue, err := queue.NewQueue(base.Context(), "repro", base.Store())
	if err != nil {
		return nil, err
	}

	prov := simple.NewProvider(
		base.Context(),
		queue,
		base.DHT(),
	)

	reprov := simple.NewReprovider(
		base.Context(),
		reprovideInterval,
		base.DHT(),
		simple.NewBlockstoreProvider(bstore),
	)

	reprovider := provider.NewSystem(prov, reprov)
	reprovider.Run()

	return reprovider, nil
}

// Session returns a session-based NodeGetter.
func Session(sp StoragePeer) ipld.NodeGetter {
	dag := sp.DagService()
	ng := merkledag.NewSession(sp.Context(), dag)
	if ng == dag {
		sp.Logger().Warn("DAGService does not support sessions")
	}
	return ng
}

func (sp *storagePeer) DagService() ipld.DAGService {
	return sp
}

func (sp *storagePeer) BlockService() blockservice.BlockService {
	return sp.bsrv
}

func (sp *storagePeer) Reprovider() provider.System {
	return sp.reprovider
}

func (sp *storagePeer) Context() context.Context {
	return sp.base.Context()
}

func (sp *storagePeer) Close() error {
	errs := []error{}

	if err := sp.reprovider.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := sp.bsrv.Close(); err != nil {
		errs = append(errs, err)
	}

	baseErr := sp.base.Close()

	if len(errs) == 0 {
		return baseErr
	}
	return errors.New("could not close ipld node")
}

func (sp *storagePeer) PrivKey() crypto.PrivKey {
	return sp.base.PrivKey()
}

func (sp *storagePeer) Psk() pnet.PSK {
	return sp.base.Psk()
}

func (sp *storagePeer) Host() host.Host {
	return sp.base.Host()
}

func (sp *storagePeer) DHT() *kaddht.IpfsDHT {
	return sp.base.DHT()
}

func (sp *storagePeer) Store() datastore.Batching {
	return sp.base.Store()
}

func (sp *storagePeer) Logger() logging.EventLogger {
	return sp.base.Logger()
}

func (sp *storagePeer) PubSub() *pubsub.PubSub {
	return sp.base.PubSub()
}

func (sp *storagePeer) Topics() map[string]*pubsub.Topic {
	return sp.base.Topics()
}
