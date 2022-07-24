package pubsub

import (
	"context"
	"sync"
	"time"

	pubsublibp2p "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
)

type PubsubService interface {
	Pubsub() *pubsublibp2p.PubSub
	Publish(topicName string, data []byte) error
	GetTopic(topicName string) *pubsublibp2p.Topic
	GetSubscription(topicName string) *pubsublibp2p.Subscription
	UnSubscribe(topicName string) error
	Subscribe(topicName string, bufferSize int) (chan *pubsublibp2p.Message, error)
}

var (
	pubsubPublishTimeout = 5 * time.Second
)

type TopicConfigurer func(topic *pubsublibp2p.Topic)

type pubsubService struct {
	ctx context.Context
	ps  *pubsublibp2p.PubSub

	topics map[string]*pubsublibp2p.Topic
	subs   map[string]*pubsublibp2p.Subscription
	lock   *sync.RWMutex

	configurer TopicConfigurer
}

func NewPubsubService(ctx context.Context, ps *pubsublibp2p.PubSub, configurer TopicConfigurer) PubsubService {
	return &pubsubService{
		ctx:        ctx,
		ps:         ps,
		topics:     make(map[string]*pubsublibp2p.Topic),
		subs:       make(map[string]*pubsublibp2p.Subscription),
		lock:       &sync.RWMutex{},
		configurer: configurer,
	}
}

func (pst *pubsubService) Pubsub() *pubsublibp2p.PubSub {
	return pst.ps
}

func (pst *pubsubService) GetTopic(topicName string) *pubsublibp2p.Topic {
	pst.lock.RLock()
	defer pst.lock.RUnlock()

	t, ok := pst.topics[topicName]
	if !ok {
		return nil
	}
	return t
}

func (pst *pubsubService) GetSubscription(topicName string) *pubsublibp2p.Subscription {
	pst.lock.RLock()
	defer pst.lock.RUnlock()

	s, ok := pst.subs[topicName]
	if !ok {
		return nil
	}
	return s
}

func (pst *pubsubService) Publish(topicName string, data []byte) error {
	fctx, cancel := context.WithTimeout(pst.ctx, pubsubPublishTimeout)
	defer cancel()
	topic := pst.GetTopic(topicName)
	if topic == nil {
		return errors.Errorf("topic not found: %s", topicName)
	}
	err := topic.Publish(fctx, data)
	if err == nil {
		metricPubsubOut.WithLabelValues(topicName).Inc()
	}
	return err
}

func (pst *pubsubService) UnSubscribe(topicName string) error {
	pst.lock.Lock()
	defer pst.lock.Unlock()

	topic, ok := pst.topics[topicName]
	if !ok {
		return nil
	}
	s, ok := pst.subs[topicName]
	if !ok {
		return nil
	}
	s.Cancel()
	err := topic.Close()

	delete(pst.topics, topicName)
	delete(pst.subs, topicName)

	return err
}

func (pst *pubsubService) Subscribe(topicName string, bufferSize int) (chan *pubsublibp2p.Message, error) {
	pst.lock.Lock()
	defer pst.lock.Unlock()

	sub, err := pst.subscribe(topicName)
	if err != nil {
		return nil, err
	}

	return pst.listen(sub, bufferSize), nil
}

func (pst *pubsubService) subscribe(topicName string) (*pubsublibp2p.Subscription, error) {
	t, ok := pst.topics[topicName]
	if !ok {
		topic, err := pst.ps.Join(topicName)
		if err != nil {
			return nil, err
		}
		pst.topics[topicName] = topic
		t = topic
	}

	s, ok := pst.subs[topicName]
	if ok && s != nil {
		// already subscribed
		return s, nil
	}
	sub, err := t.Subscribe()
	if err != nil {
		_ = t.Close()
		delete(pst.topics, topicName)
		return nil, err
	}
	pst.subs[topicName] = sub

	return sub, nil
}

func (pst *pubsubService) listen(sub *pubsublibp2p.Subscription, bufferSize int) chan *pubsublibp2p.Message {
	receiver := make(chan *pubsublibp2p.Message, bufferSize)

	go func() {
		topicName := sub.Topic()
		ctx, cancel := context.WithCancel(pst.ctx)
		defer func() {
			metricPubsubListening.WithLabelValues(topicName).Dec()
			close(receiver)
			sub.Cancel()
			cancel()
		}()

		metricPubsubListening.WithLabelValues(topicName).Inc()
		for ctx.Err() == nil {
			next, err := sub.Next(ctx)
			if err != nil {
				switch err {
				case pubsublibp2p.ErrSubscriptionCancelled, pubsublibp2p.ErrTopicClosed:
					// subscription was destroyed > exit
					return
				default:
				}
				continue
			}
			if next == nil {
				continue
			}
			select {
			case receiver <- next:
				metricPubsubIn.WithLabelValues(topicName).Inc()
			default:
				metricPubsubInDropped.WithLabelValues(topicName).Inc()
				// dropping message as channel is full
			}
		}
	}()

	return receiver
}