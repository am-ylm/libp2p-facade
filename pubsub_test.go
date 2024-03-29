package p2pfacade

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	logging "github.com/ipfs/go-log/v2"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/stretchr/testify/require"
)

func TestPubsub(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, logging.SetLogLevelRegex("p2p:.*", "debug"))
	n := 10
	nodes := newLocalNetwork(ctx, t, n)

	<-time.After(time.Second)

	topicName := "mytest"
	var wg sync.WaitGroup
	var msgCount int64
	for _, f := range nodes {
		wg.Add(1)
		go func(facade Facade) {
			err := facade.Subscribe(topicName, func(msg *pubsub.Message) {
				require.NotNil(t, msg)
				atomic.AddInt64(&msgCount, 1)
			}, 2)
			require.NoError(t, err)
			wg.Done()
		}(f)
	}

	wg.Wait()
	t.Log("nodes are subscribed to topic", topicName)

	// ensure that we have enough peers
	for _, f := range nodes {
		topic := f.GetTopic(topicName)
		require.NotNil(t, topic)
		peers := topic.ListPeers()
		require.GreaterOrEqual(t, len(peers), 2)
	}

	for i, f := range nodes {
		wg.Add(1)
		go func(facade Facade, i int) {
			defer wg.Done()
			<-time.After(time.Duration(rand.Intn(10)+1) * time.Millisecond * 50)
			data := []byte(fmt.Sprintf("msg-from-node-%d", i))
			require.NoError(t, facade.Publish(topicName, data))
		}(f, i)
	}

	wg.Wait()

	<-time.After(time.Second * 2)

	require.GreaterOrEqual(t, atomic.LoadInt64(&msgCount), int64(n))

	for _, f := range nodes {
		require.NoError(t, f.Close())
	}
}
