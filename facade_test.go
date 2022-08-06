package p2pfacade

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amirylm/libp2p-facade/streams"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/stretchr/testify/require"
)

// TODO: remove timeouts
func TestStreams(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := 10
	nodes := newLocalNetwork(ctx, t, n)

	<-time.After(2 * time.Second)

	pid := core.ProtocolID("/mytest")
	var wg sync.WaitGroup
	var successMsgCount int64
	for _, f := range nodes {
		wg.Add(1)
		go func(facade Facade) {
			defer wg.Done()
			facade.Host().SetStreamHandler(pid, func(s core.Stream) {
				req, res, done, err := streams.HandleStream(s, 5*time.Second)
				require.NoError(t, err)
				defer func() {
					_ = done()
				}()
				require.True(t, strings.Contains(string(req), "msg-from-node"))
				require.NoError(t, res(req))
				atomic.AddInt64(&successMsgCount, 1)
			})
		}(f)
	}
	wg.Wait()
	<-time.After(2 * time.Second)
	t.Log("configured stream handlers", pid)
	send := func(from, to Facade, i int) {
		defer wg.Done()
		res, err := streams.Request(to.Host().ID(), pid, []byte(fmt.Sprintf("msg-from-node-%d", i)), streams.StreamConfig{
			Ctx:     ctx,
			Host:    from.Host(),
			Timeout: time.Second * 2,
		})
		require.NoError(t, err)
		require.True(t, strings.Contains(string(res), "msg-from-node"))
	}

	wg.Add(4)
	go func() {
		go send(nodes[0], nodes[1], 0)
		go send(nodes[2], nodes[5], 2)
	}()
	go send(nodes[0], nodes[3], 0)
	go send(nodes[5], nodes[4], 5)

	wg.Wait()

	<-time.After(2 * time.Second)

	require.GreaterOrEqual(t, atomic.LoadInt64(&successMsgCount), int64(3))

	for _, f := range nodes {
		require.NoError(t, f.Close())
	}

	t.Logf("done")
}
