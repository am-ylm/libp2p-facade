package commons

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

// EnsureConnectedPeers ensures that the host has at least {n} connected peers under the given timeout
func EnsureConnectedPeers(pctx context.Context, h host.Host, n int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(pctx, timeout)
	defer cancel()

	var connectedPeers []peer.ID
	for ctx.Err() == nil {
		connectedPeers = h.Network().Peers()
		if len(connectedPeers) >= n {
			return nil
		}
	}

	return fmt.Errorf("found only %d connected peers, expected %d", len(connectedPeers), n)
}
