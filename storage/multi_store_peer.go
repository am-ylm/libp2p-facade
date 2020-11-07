package storage

import (
	"github.com/ipfs/go-datastore"
	crdt "github.com/ipfs/go-ds-crdt"
)

var (
	StoreMain = "store_main"
)

type MultiStorePeer struct {
	StoragePeer

	crdts map[string]*crdt.Datastore

	stores map[string]datastore.Batching
}

func NewMultiStorePeer(sp StoragePeer) *MultiStorePeer {
	msp := MultiStorePeer{sp, map[string]*crdt.Datastore{}, map[string]datastore.Batching{}}

	msp.UseDatastore(StoreMain, sp.Store())

	return &msp
}

func (msp *MultiStorePeer) Crdt(name string) *crdt.Datastore {
	return msp.crdts[name]
}

func (msp *MultiStorePeer) UseCrdt(name string, store *crdt.Datastore) bool {
	if _, exists := msp.crdts[name]; !exists {
		msp.crdts[name] = store
		return true
	}
	return false
}

func (msp *MultiStorePeer) Datastore(name string) datastore.Batching {
	return msp.stores[name]
}

func (msp *MultiStorePeer) UseDatastore(name string, store datastore.Batching) bool {
	if _, exists := msp.stores[name]; !exists {
		msp.stores[name] = store
		return true
	}
	return false
}
