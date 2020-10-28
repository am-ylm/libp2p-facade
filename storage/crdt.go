package storage

import (
	"github.com/amirylm/libp2p-facade/core"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	crdt "github.com/ipfs/go-ds-crdt"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	ipld "github.com/ipfs/go-ipld-format"
)

func ConfigureCrdt(sp StoragePeer, name string, crdtOpts *crdt.Options) (*crdt.Datastore, error) {
	if crdtOpts == nil {
		crdtOpts = crdt.DefaultOptions()
	}
	// always override logger
	crdtOpts.Logger = sp.Logger()

	pubsubBC, err := crdt.NewPubSubBroadcaster(sp.Context(), sp.PubSub(), name)
	if err != nil {
		sp.Logger().Fatalf("could not create crdt.PubSubBroadcaster: %s", err.Error())
	}

	dsyncer := NewDagSyncer(sp.DagService(), sp.BlockService().Blockstore())
	crdt, err := crdt.New(sp.Store(), ds.NewKey(name), dsyncer, pubsubBC, crdtOpts)
	if err != nil {
		sp.Logger().Fatal(err)
	}
	go core.AutoClose(sp.Context(), crdt)

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
