package pubsub

import (
	"github.com/amirylm/libp2p-facade/config"
	pubsublibp2p "github.com/libp2p/go-libp2p-pubsub"
)

type nilConfigurer struct{}

// Opts implements Configurer
func (*nilConfigurer) Opts() []pubsublibp2p.Option {
	return nil
}

// PubOpts implements Configurer
func (*nilConfigurer) PubOpts(topicName string) []pubsublibp2p.PubOpt {
	return nil
}

// SubOpts implements Configurer
func (*nilConfigurer) SubOpts(topicName string) []pubsublibp2p.SubOpt {
	return nil
}

// Topic implements Configurer
func (*nilConfigurer) Topic(topic *pubsublibp2p.Topic) {
}

// TopicOpts implements Configurer
func (*nilConfigurer) TopicOpts(topicName string) []pubsublibp2p.TopicOpt {
	return nil
}

// TopicValidator implements Configurer
func (*nilConfigurer) TopicValidator(topicName string) (pubsublibp2p.ValidatorEx, []pubsublibp2p.ValidatorOpt) {
	return nil, nil
}

// NewNilConfigurer creates an empty config.PubsubConfigurer
func NewNilConfigurer() config.PubsubConfigurer {
	return &nilConfigurer{}
}
