package config

import (
	crand "crypto/rand"
	"encoding/json"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/libp2p/go-libp2p-core/routing"
	pubsublibp2p "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	yamuxID = "/yamux/1.0.0"
)

// StaticConfig contains static configuration for the p2p node
type StaticConfig struct {
	// ListenAddrs addrs to listen, this allows to specify also the transports that are supported
	ListenAddrs []string `json:"listenAddrs" yaml:"listenAddrs"`
	// Relayers are possible circuit relay end-points
	Relayers []string `json:"relayers,omitempty" yaml:"relayers,omitempty"`
	// DialTimeout is the timeout to use when dialing peers
	DialTimeout time.Duration `json:"dialTimeout,omitempty" yaml:"dialTimeout,omitempty"`
	// Muxers the supported muxers
	Muxers []string `json:"muxers,omitempty" yaml:"muxers,omitempty"`
	// Security the supported security protocols
	Security []string `json:"security,omitempty" yaml:"security,omitempty"`
	// NetworkSecret is a secret to use for a private network
	NetworkSecret string `json:"networkSecret,omitempty" yaml:"networkSecret,omitempty"`
	// DisablePing is a negative flag to turn off libp2p Ping
	DisablePing bool `json:"disablePing,omitempty" yaml:"disablePing,omitempty"`
	// EnableAutoRelay whether to enable auto relay
	EnableAutoRelay bool `json:"enableAutoRelay,omitempty" yaml:"enableAutoRelay,omitempty"`
	// MdnsServiceTag is the service tag used by mdns service. mdns is disabled if service tag is empty
	MdnsServiceTag string `json:"mdnsServiceTag,omitempty" yaml:"mdnsServiceTag,omitempty"`
	// UserAgent is the user agent string used by identify protocol
	UserAgent string `json:"userAgent,omitempty" yaml:"userAgent,omitempty"`
}

// Config contains both dynamic (libp2p components) and static information (json/yaml).
type Config struct {
	// StaticConfig
	StaticConfig
	// PrivateKey is used as the private key of the peer
	PrivateKey crypto.PrivKey
	// Routing configures routing.Routing for the given host
	Routing func(h host.Host) (routing.Routing, error)
	// PubsubConfigurer enables to configure pubsub components dynamically
	PubsubConfigurer PubsubConfigurer
	// Opts is used to inject own options
	Opts []libp2p.Option
}

// UnmarshalJSON implements json.Unmarshaler
func (c *Config) UnmarshalJSON(b []byte) error {
	var static StaticConfig
	if err := json.Unmarshal(b, &static); err != nil {
		return err
	}
	c.StaticConfig = static

	return nil
}

// MarshalJSON implements json.Marshaler
func (c *Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(&c.StaticConfig)
}

// UnmarshalYAML implements yaml.Unmarshaler
func (c *Config) UnmarshalYAML(b []byte) error {
	var static StaticConfig
	if err := yaml.Unmarshal(b, &static); err != nil {
		return err
	}
	c.StaticConfig = static

	return nil
}

// MarshalYAML implements yaml.Marshaler
func (c *Config) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(&c.StaticConfig)
}

// Libp2pOptions returns a list of libp2p options for the given config
func (cfg *Config) Libp2pOptions() ([]libp2p.Option, error) {
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(cfg.ListenAddrs...),
		libp2p.Identity(cfg.PrivateKey),
	}

	for _, sec := range cfg.Security {
		switch sec {
		case noise.ID:
			opts = append(opts, libp2p.Security(noise.ID, noise.New))
		case libp2ptls.ID:
			opts = append(opts, libp2p.Security(libp2ptls.ID, libp2ptls.New))
		default:
		}
	}

	for _, sec := range cfg.Muxers {
		switch sec {
		case yamuxID:
			opts = append(opts, libp2p.Muxer(yamuxID, yamux.DefaultTransport))
		default:
		}
	}

	if len(cfg.NetworkSecret) > 0 {
		opts = append(opts, libp2p.PrivateNetwork(pnet.PSK(cfg.NetworkSecret)))
	}

	if cfg.DialTimeout > 0 {
		opts = append(opts, libp2p.WithDialTimeout(cfg.DialTimeout))
	}

	opts = append(opts, libp2p.Ping(!cfg.DisablePing))

	if len(cfg.UserAgent) > 0 {
		opts = append(opts, libp2p.UserAgent(cfg.UserAgent))
	}

	if len(cfg.Relayers) > 0 {
		opts = append(opts, libp2p.EnableRelay())
		rels := make([]peer.AddrInfo, 0)
		for _, rel := range cfg.Relayers {
			pi, err := peer.AddrInfoFromString(rel)
			if err != nil {
				continue
			}
			rels = append(rels, *pi)
		}
		opts = append(opts, libp2p.StaticRelays(rels))
	}
	if cfg.EnableAutoRelay {
		opts = append(opts, libp2p.EnableAutoRelay())
	}

	opts = append(opts, cfg.Opts...)

	return opts, nil
}

// Init initialize config
func (cfg *Config) Init() error {
	if err := cfg.initPrivateKey(); err != nil {
		return errors.Wrap(err, "could not initialize private key")
	}
	if len(cfg.Security) == 0 {
		// using noise security by default
		cfg.Security = []string{noise.ID}
	}
	if len(cfg.Muxers) == 0 {
		// using yamux muxer by default
		cfg.Muxers = []string{yamuxID}
	}
	return nil
}

func (cfg *Config) initPrivateKey() error {
	if cfg.PrivateKey == nil {
		sk, _, err := crypto.GenerateECDSAKeyPair(crand.Reader)
		if err != nil {
			return err
		}
		cfg.PrivateKey = sk
	}
	return nil
}

// PubsubConfigurer helps to aid in a custom set of configurations for pubsub
type PubsubConfigurer interface {
	// Topic enalbes to configure a topic, e.g. score params
	Topic(topic *pubsublibp2p.Topic)
	// TopicValidator returns the topic validator of this topic
	TopicValidator(topicName string) (pubsublibp2p.ValidatorEx, []pubsublibp2p.ValidatorOpt)
	// Opts enables to inject any set of pubsublibp2p.Option
	Opts() []pubsublibp2p.Option
	// TopicsOpts returns the opts for the given topic
	TopicOpts(topicName string) []pubsublibp2p.TopicOpt
	// SubOpts are the opts for a subscription of the given topic
	SubOpts(topicName string) []pubsublibp2p.SubOpt
	// PubOpts is the publish options
	PubOpts(topicName string) []pubsublibp2p.PubOpt
}
