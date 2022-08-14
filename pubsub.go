package p2pfacade

import (
	"github.com/amirylm/libp2p-facade/pubsub"
	pubsublibp2p "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
)

func (f *facade) setupPubsub() error {
	if f.cfg.PubsubConfigurer == nil {
		return nil
	}
	opts := make([]pubsublibp2p.Option, 0)
	opts = append(opts, pubsublibp2p.WithEventTracer(pubsub.NewReportingTracer()))
	opts = append(opts, f.cfg.PubsubConfigurer.Opts()...)
	ps, err := pubsublibp2p.NewGossipSub(f.ctx, f.host, opts...)
	if err != nil {
		return errors.Wrap(err, "could not setup pubsub")
	}
	f.ps = pubsub.NewPubsubService(f.ctx, ps, f.cfg.PubsubConfigurer)

	return nil
}

// GetSubscription implements Facade
func (f *facade) GetSubscription(topicName string) *pubsublibp2p.Subscription {
	return f.ps.GetSubscription(topicName)
}

// GetTopic implements Facade
func (f *facade) GetTopic(topicName string) *pubsublibp2p.Topic {
	return f.ps.GetTopic(topicName)
}

// Pubsub implements Facade
func (f *facade) Pubsub() *pubsublibp2p.PubSub {
	return f.ps.Pubsub()
}

// Publish implements Facade
func (f *facade) Publish(topicName string, data []byte) error {
	return f.ps.Publish(topicName, data)
}

// Subscribe implements Facade
func (f *facade) Subscribe(topicName string, handler pubsub.PubsubHandler, bufferSize int) error {
	// topicCfgs := f.cfg.Pubsub.GetTopicCfg(topicName)
	// if len(topicCfgs) > 0 {
	// 	topicCfg := topicCfgs[0] // TODO: consider other configs
	// 	if topicCfg.BufferSize > 0 {
	// 		bufferSize = topicCfg.BufferSize
	// 	}
	// }
	return f.ps.Subscribe(topicName, handler, bufferSize)
}

// UnSubscribe implements Facade
func (f *facade) UnSubscribe(topicName string) error {
	return f.ps.UnSubscribe(topicName)
}
