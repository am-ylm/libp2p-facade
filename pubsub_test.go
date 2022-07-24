package p2pfacade

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPubsub(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := 10
	nodes := newLocalNetwork(ctx, t, n)

	<-time.After(time.Second)

	topicName := "mytest"
	var wg sync.WaitGroup
	var msgCount int64
	for _, f := range nodes {
		wg.Add(1)
		go func(facade Facade) {
			cn, err := facade.Subscribe(topicName, 2)
			require.NoError(t, err)
			wg.Done()
			for msg := range cn {
				require.NotNil(t, msg)
				atomic.AddInt64(&msgCount, 1)
			}
		}(f)
		require.NoError(t, f.Close())
	}

	wg.Wait()
	t.Log("nodes are subscribed to topic", topicName)

	// ensure that we have enough peers
	for _, f := range nodes {
		topic := f.GetTopic(topicName)
		require.NotNil(t, topic)
		peers := topic.ListPeers()
		require.GreaterOrEqual(t, 2, len(peers))
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
