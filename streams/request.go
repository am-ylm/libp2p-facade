package streams

import (
	"context"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/pkg/errors"
)

const (
	DefaultTimeout = 15 * time.Second
)

var (
	logger = logging.Logger("p2p:stream")
)

// StreamConfig is the config object required to make a request
type StreamConfig struct {
	Ctx     context.Context
	Host    host.Host
	Timeout time.Duration
}

// Request sends a message to the given stream and returns the response
func Request(peerID peer.ID, protocol protocol.ID, data []byte, cfg StreamConfig) ([]byte, error) {
	s, err := cfg.Host.NewStream(cfg.Ctx, peerID, protocol)
	if err != nil {
		return nil, err
	}
	metricStreamOut.WithLabelValues(string(protocol)).Inc()
	stream := NewStream(s)
	defer func() {
		_ = stream.Close()
	}()
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	if err := stream.WriteWithTimeout(data, timeout); err != nil {
		metricStreamOutDone.WithLabelValues(string(protocol), "write").Inc()
		return nil, errors.Wrap(err, "could not write to stream")
	}
	if err := s.CloseWrite(); err != nil {
		metricStreamOutDone.WithLabelValues(string(protocol), "close_write").Inc()
		return nil, errors.Wrap(err, "could not close-write stream")
	}
	res, err := stream.ReadWithTimeout(timeout)
	if err != nil {
		metricStreamOutDone.WithLabelValues(string(protocol), "read").Inc()
		return nil, errors.Wrap(err, "could not read stream msg")
	}
	logger.Debugf("successful stream request %s, target peer: %s", string(protocol), peerID.String())
	metricStreamOutDone.WithLabelValues(string(protocol), "").Inc()
	return res, nil
}
