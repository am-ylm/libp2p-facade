package lib

import (
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type PubsubEmitter struct {
	ps    *pubsub.PubSub

	self peer.ID

	topics map[string]*pubsub.Topic
}

func NewPubSubEmitter(ps *pubsub.PubSub, self peer.ID) *PubsubEmitter {
	topics := map[string]*pubsub.Topic{}

	em := PubsubEmitter{ps, self, topics}

	return &em
}

func (em *PubsubEmitter) Topic(name string) (*pubsub.Topic, error) {
	_, exist := em.topics[name]
	if !exist {
		topic, err := em.ps.Join(name)
		if err != nil {
			return nil, err
		}
		em.topics[name] = topic
	}
	return em.topics[name], nil
}

func (em *PubsubEmitter) Subscribe(name string) (*pubsub.Subscription, error) {
	topic, err := em.Topic(name)
	if err != nil {
		return nil, err
	}
	return topic.Subscribe()
}