package p2pfacade

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/amirylm/libp2p-facade/commons"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/stretchr/testify/require"
)

func newLocalNetwork(ctx context.Context, t *testing.T, n int) []Facade {
	cfgs := []*commons.Config{}
	for i := 0; i < n; i++ {
		cfgs = append(cfgs, newLocalConfig(ctx, i, n))
	}
	nodes, err := StartNodes(ctx, cfgs)
	require.NoError(t, err)
	require.Len(t, nodes, 10)

	for _, f := range nodes {
		require.NoError(t, commons.EnsureConnectedPeers(ctx, f.Host(), n/2, time.Second*8))
	}
	<-time.After(time.Second)
	return nodes
}

func newLocalConfig(ctx context.Context, i, maxPeers int) *commons.Config {
	cfg := commons.Config{
		Pubsub: &commons.PubsubConfig{
			Config: &commons.PubsubTopicConfig{},
			Topics: []commons.PubsubTopicConfig{},
		},
		Routing: func(h host.Host) (routing.Routing, error) {
			kad, _, err := NewKadDHT(ctx, h, "test.mdns", dht.ModeAutoServer, nil)
			return kad, err
		},
	}
	cfg.ListenAddrs = []string{"/ip4/0.0.0.0/tcp/0"}
	cfg.UserAgent = fmt.Sprintf("test/v0/%d", i)
	cfg.MdnsServiceTag = "test.mdns"
	return &cfg
}
