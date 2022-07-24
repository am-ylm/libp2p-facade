package streams

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/pkg/errors"
)

const (
	DefaultTimeout = 15 * time.Second
)

// StreamConfig is the config object required to make a request
type StreamConfig struct {
	Ctx     context.Context
	Host    host.Host
	Timeout time.Duration
}

// Request sends a message to the given stream and returns the response
func Request(peerID peer.ID, protocol protocol.ID, data []byte, cfg StreamConfig) ([]byte, error) {
	metricStreamOutActive.WithLabelValues(string(protocol)).Inc()
	defer metricStreamOutActive.WithLabelValues(string(protocol)).Dec()
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
		metricStreamOutFailed.WithLabelValues(string(protocol), "write").Inc()
		return nil, errors.Wrap(err, "could not write to stream")
	}
	if err := s.CloseWrite(); err != nil {
		metricStreamOutFailed.WithLabelValues(string(protocol), "close_write").Inc()
		return nil, errors.Wrap(err, "could not close-write stream")
	}
	res, err := stream.ReadWithTimeout(timeout)
	if err != nil {
		metricStreamOutFailed.WithLabelValues(string(protocol), "read").Inc()
		return nil, errors.Wrap(err, "could not read stream msg")
	}
	metricStreamOutSuccess.WithLabelValues(string(protocol)).Inc()
	return res, nil
}
