package streams

import (
	"time"

	core "github.com/libp2p/go-libp2p-core"
	"github.com/pkg/errors"
)

// RespondStream abstracts stream with a function that accepts only the data to send
type RespondStream func([]byte) error

// CloseStream closes the stream
type CloseStream func() error

// HandleStream is called at the beginning of stream handlers to create a wrapper stream and read first message
func HandleStream(stream core.Stream, timeout time.Duration) ([]byte, RespondStream, CloseStream, error) {
	protocol := stream.Protocol()

	metricStreamIn.WithLabelValues(string(protocol)).Inc()
	s := NewStream(stream)
	pid := stream.Conn().RemotePeer().String()
	done := func() error {
		logger.Debugf("closing stream %s, src peer: %s", string(protocol), pid)
		return s.Close()
	}
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	data, err := s.ReadWithTimeout(timeout)
	if err != nil {
		metricStreamInDone.WithLabelValues(string(protocol), "read").Inc()
		return nil, nil, done, errors.Wrap(err, "could not read stream msg")
	}
	respond := func(res []byte) error {
		if err := s.WriteWithTimeout(res, timeout); err != nil {
			metricStreamInDone.WithLabelValues(string(protocol), "write").Inc()
			return errors.Wrap(err, "could not write to stream")
		}
		logger.Debugf("handle stream success %s, src peer: %s", string(protocol), pid)
		metricStreamInDone.WithLabelValues(string(protocol), "").Inc()
		return nil
	}

	return data, respond, done, nil
}
