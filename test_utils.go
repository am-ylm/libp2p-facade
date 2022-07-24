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
	nodes := []Facade{}
	t.Logf("creating %d nodes", n)
	for i := 0; i < n; i++ {
		f, err := New(ctx, newLocalConfig(i, n))
		require.NoError(t, err)
		nodes = append(nodes, f)

		go func(facade Facade) {
			require.NoError(t, facade.Start(nil))
		}(f)
	}

	require.Len(t, nodes, 10)

	for _, f := range nodes {
		require.NoError(t, commons.EnsureConnectedPeers(ctx, f.Host(), n-2, time.Second*5))
	}

	t.Log("nodes are connected")

	<-time.After(time.Second)
	return nodes
}

func newLocalConfig(i, maxPeers int) *commons.Config {
	cfg := commons.Config{
		// MaxPeers:    maxPeers,
		// Discovery: []p2pconfig.DiscoveryCfg{
		// 	{
		// 		Type:       mdns.Name,
		// 		ServiceTag: "test.mdns.v0",
		// 	},
		// 	{
		// 		Type:       kaddht.Name,
		// 		ServiceTag: "test.kaddht.v0",
		// 	},
		// },
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
