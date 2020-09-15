package ipld

import (
	"github.com/amirylm/priv-libp2p-node/core"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	crdt "github.com/ipfs/go-ds-crdt"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	ipld "github.com/ipfs/go-ipld-format"
)

func ConfigureCrdt(node IpldNode, topicName string, crdtOpts *crdt.Options) (*crdt.Datastore, error) {
	if crdtOpts == nil {
		crdtOpts = crdt.DefaultOptions()
	}
	// always override logger
	crdtOpts.Logger = node.Logger()

	pubsubBC, err := crdt.NewPubSubBroadcaster(node.Context(), node.PubSub(), topicName)
	if err != nil {
		node.Logger().Fatalf("could not create crdt.PubSubBroadcaster: %s", err.Error())
	}

	dsync := NewDagSyncer(node.DagService(), node.BlockService().Blockstore())
	crdt, err := crdt.New(node.Store(), ds.NewKey("crdt"), dsync, pubsubBC, crdtOpts)
	if err != nil {
		node.Logger().Fatal(err)
	}
	go core.AutoClose(node.Context(), crdt)

	return crdt, err
}

func NewDagSyncer(base ipld.DAGService, bs blockstore.Blockstore) *dagSyncer {
	n := dagSyncer{base, bs}

	return &n
}

type dagSyncer struct {
	ipld.DAGService

	bs blockstore.Blockstore
}

func (n dagSyncer) HasBlock(c cid.Cid) (bool, error) {
	return n.bs.Has(c)
}

