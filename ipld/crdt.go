package ipld

import (
	"github.com/amirylm/priv-libp2p-node/core"
	ds "github.com/ipfs/go-datastore"
	crdt "github.com/ipfs/go-ds-crdt"
)

type CrdtConfig struct {
	TopicName string
}

func ConfigureCrdt(node IpldNode, crdtOpts *crdt.Options, topicName string) (*crdt.Datastore, error) {
	if crdtOpts == nil {
		crdtOpts = crdt.DefaultOptions()
	}
	// always override logger
	crdtOpts.Logger = node.Logger()

	pubsubBC, err := crdt.NewPubSubBroadcaster(node.Context(), node.PubSub(), topicName)
	if err != nil {
		node.Logger().Fatalf("could not create crdt.PubSubBroadcaster: %s", err.Error())
	}

	crdt, err := crdt.New(node.Store(), ds.NewKey("crdt"), node.(*ipldNode), pubsubBC, crdtOpts)
	if err != nil {
		node.Logger().Fatal(err)
	}
	go core.AutoClose(node.Context(), crdt)

	return crdt, err
}