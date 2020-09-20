package core

import (
	"context"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type PubSuber interface {
	PubSub() *pubsub.PubSub
	Topics() map[string]*pubsub.Topic
}

func Topic(node PubSuber, name string) (*pubsub.Topic, error) {
	topics := node.Topics()
	_, exist := topics[name]
	if !exist {
		topic, err := node.PubSub().Join(name)
		if err != nil {
			return nil, err
		}
		topics[name] = topic
	}
	return topics[name], nil
}

func Subscribe(node PubSuber, topic string) (*pubsub.Subscription, error) {
	t, err := Topic(node, topic)
	if err != nil {
		return nil, err
	}
	return t.Subscribe()
}

func Publish(node PubSuber, ctx context.Context, topic string, data []byte) error {
	t, err := Topic(node, topic)
	if err != nil {
		return err
	}
	return t.Publish(ctx, data)
}
