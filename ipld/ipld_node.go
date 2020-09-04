package ipld

import (
	"context"
	"github.com/amirylm/priv-libp2p-node/core"
	"github.com/ipfs/go-bitswap"
	"github.com/ipfs/go-bitswap/network"
	"github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	exoffline "github.com/ipfs/go-ipfs-exchange-offline"
	"github.com/ipfs/go-ipfs-provider/queue"
	"github.com/ipfs/go-ipfs-provider/simple"
	ipld "github.com/ipfs/go-ipld-format"
	provider "github.com/ipfs/go-ipfs-provider"
	"github.com/ipfs/go-blockservice"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipfs/go-merkledag"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/pnet"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"log"
	"time"
)

const (
	defaultReprovideInterval = 8 * time.Hour
)

type IpldNode interface {
	core.LibP2PNode

	DAGService() ipld.DAGService
	BlockService() blockservice.BlockService
	Reprovider() provider.System
}

type ipldNode struct {
	base core.LibP2PNode

	bsrv       blockservice.BlockService
	dag        ipld.DAGService
	reprovider provider.System
}

func NewIpldNode(base core.LibP2PNode, offline bool) IpldNode {
	dag, bsrv, bstore, err := createDagServices(base, offline)
	if err != nil {
		log.Panic("could not create DAG services")
	}
	repro, err := setupReprovider(base, bstore, defaultReprovideInterval, offline)
	node := ipldNode{base, bsrv, dag, repro}
	return &node
}

func createDagServices(base core.LibP2PNode, offline bool) (ipld.DAGService, blockservice.BlockService, blockstore.Blockstore, error) {
	var bsrv blockservice.BlockService
	var bs blockstore.Blockstore

	bs = blockstore.NewBlockstore(base.Store())
	bs = blockstore.NewIdStore(bs)
	bs, _ = blockstore.CachedBlockstore(base.Context(), bs, blockstore.DefaultCacheOpts())
	if offline {
		bsrv = blockservice.New(bs, exoffline.Exchange(bs))
		return merkledag.NewDAGService(bsrv), bsrv, bs, nil
	}

	bswapnet := network.NewFromIpfsHost(base.Host(), base.DHT())
	bswap := bitswap.New(base.Context(), bswapnet, bs)
	bsrv = blockservice.New(bs, bswap)

	return merkledag.NewDAGService(bsrv), bsrv, bs, nil
}

func setupReprovider(base core.LibP2PNode, bstore blockstore.Blockstore, reprovideInterval time.Duration, offline bool) (provider.System, error) {
	if offline {
		return provider.NewOfflineProvider(), nil
	}

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
func Session(node IpldNode) ipld.NodeGetter {
	ng := merkledag.NewSession(node.Context(), node.DAGService())
	if ng == node.DAGService() {
		node.Logger().Warn("DAGService does not support sessions")
	}
	return ng
}

func (n *ipldNode) DAGService() ipld.DAGService {
	return n.dag
}

func (n *ipldNode) BlockService() blockservice.BlockService {
	return n.bsrv
}

func (n *ipldNode) Reprovider() provider.System {
	return n.reprovider
}

func (n *ipldNode) Context() context.Context {
	return n.base.Context()
}

func (n *ipldNode) Close() error {
	return n.base.Close()
}

func (n *ipldNode) PrivKey() crypto.PrivKey {
	return n.base.PrivKey()
}

func (n *ipldNode) Psk() pnet.PSK {
	return n.base.Psk()
}

func (n *ipldNode) Host() host.Host {
	return n.base.Host()
}

func (n *ipldNode) DHT() *kaddht.IpfsDHT {
	return n.base.DHT()
}

func (n *ipldNode) Store() datastore.Batching {
	return n.base.Store()
}

func (n *ipldNode) Logger() logging.EventLogger {
	return n.base.Logger()
}

func (n *ipldNode) PubSub() *pubsub.PubSub {
	return n.base.PubSub()
}

func (n *ipldNode) Topics() map[string]*pubsub.Topic {
	return n.base.Topics()
}
