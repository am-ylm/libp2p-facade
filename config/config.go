package config

import (
	crand "crypto/rand"
	"encoding/json"
	"regexp"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/libp2p/go-libp2p-core/routing"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
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
	// Pubsub is the pubsub config
	Pubsub *PubsubConfig `json:"pubsub,omitempty" yaml:"pubsub,omitempty"`
}

// Config contains both dynamic (libp2p components) and static information (json/yaml).
// TODO: complete more fields from https://pkg.go.dev/github.com/libp2p/go-libp2p/config#Config
type Config struct {
	StaticConfig
	// PrivateKey is used as the private key of the peer
	PrivateKey crypto.PrivKey
	// Reporter to use to collect metrics
	Reporter        metrics.Reporter
	AddrsFactory    bhost.AddrsFactory
	ConnectionGater connmgr.ConnectionGater
	ConnManager     connmgr.ConnManager
	ResourceManager network.ResourceManager
	Peerstore       peerstore.Peerstore

	Routing func(h host.Host) (routing.Routing, error)
}

func (c *Config) UnmarshalJSON(b []byte) error {
	var static StaticConfig
	if err := json.Unmarshal(b, &static); err != nil {
		return err
	}
	c.StaticConfig = static

	return nil
}

func (c *Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(&c.StaticConfig)
}

func (c *Config) UnmarshalYAML(b []byte) error {
	var static StaticConfig
	if err := yaml.Unmarshal(b, &static); err != nil {
		return err
	}
	c.StaticConfig = static

	return nil
}

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

	if cfg.AddrsFactory != nil {
		opts = append(opts, libp2p.AddrsFactory(cfg.AddrsFactory))
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
	if cfg.Peerstore != nil {
		opts = append(opts, libp2p.Peerstore(cfg.Peerstore))
	}
	if cfg.ConnectionGater != nil {
		opts = append(opts, libp2p.ConnectionGater(cfg.ConnectionGater))
	}
	if cfg.ConnManager != nil {
		opts = append(opts, libp2p.ConnectionManager(cfg.ConnManager))
	}
	if cfg.Reporter != nil {
		opts = append(opts, libp2p.BandwidthReporter(cfg.Reporter))
	}
	if cfg.ResourceManager != nil {
		opts = append(opts, libp2p.ResourceManager(cfg.ResourceManager))
	}

	return opts, nil
}

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

type PubsubConfig struct {
	// Config is the general configuration in the pubsub router level
	Config *PubsubTopicConfig `json:"config" yaml:"config"`
	// Topics is the configuration for topics
	Topics []PubsubTopicConfig `json:"topics" yaml:"topics"`
}

// GetTopicCfg returns all the relevant configs (including regex) for the given topic name
func (pcfg *PubsubConfig) GetTopicCfg(topicName string) []PubsubTopicConfig {
	if pcfg == nil || len(pcfg.Topics) == 0 {
		return nil
	}
	var res []PubsubTopicConfig
	for _, topicCfg := range pcfg.Topics {
		re, err := regexp.Compile(topicCfg.Name)
		if err != nil {
			continue
		}
		if re.MatchString(topicName) {
			res = append(res, topicCfg)
		}
	}
	return res
}

type PubsubTopicConfig struct {
	Name               string              `json:"name" yaml:"name"`
	BufferSize         int                 `json:"bufferSize,omitempty" yaml:"bufferSize,omitempty"`
	SubscriptionFilter *SubscriptionFilter `json:"subscriptionFilter,omitempty" yaml:"subscriptionFilter,omitempty"`
	// MsgValidator string    `json:"msgValidator" yaml:"msgValidator"`
	// TODO: add more fields
}

type SubscriptionFilter struct {
	Pattern string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Limit   int    `json:"limit,omitempty" yaml:"limit,omitempty"`
}
