package p2pfacade

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/amirylm/libp2p-facade/commons"
	"github.com/stretchr/testify/require"
)

func newLocalNetwork(ctx context.Context, t *testing.T, n int) []Facade {
	cfgs := []*commons.Config{}
	for i := 0; i < n; i++ {
		cfgs = append(cfgs, newLocalConfig(i, n))
	}
	nodes, err := StartNodes(ctx, cfgs)
	require.NoError(t, err)
	require.Len(t, nodes, 10)

	for _, f := range nodes {
		require.NoError(t, commons.EnsureConnectedPeers(ctx, f.Host(), n-2, time.Second*5))
	}
	<-time.After(time.Second)
	return nodes
}

func newLocalConfig(i, maxPeers int) *commons.Config {
	cfg := commons.Config{
		Pubsub: &commons.PubsubConfig{
			Config: &commons.PubsubTopicConfig{},
			Topics: []commons.PubsubTopicConfig{},
		},
	}
	cfg.ListenAddrs = []string{"/ip4/0.0.0.0/tcp/0"}
	cfg.UserAgent = fmt.Sprintf("test/v0/%d", i)
	cfg.MdnsServiceTag = "test.mdns.v0"
	return &cfg
}
