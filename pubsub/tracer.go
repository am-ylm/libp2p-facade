package pubsub

import (
	pubsublibp2p "github.com/libp2p/go-libp2p-pubsub"
	ps_pb "github.com/libp2p/go-libp2p-pubsub/pb"
)

// psTracer helps to trace pubsub events
type psTracer struct {
}

// newTracer creates an instance of psTracer
func NewReportingTracer() pubsublibp2p.EventTracer {
	return &psTracer{}
}

// Trace handles events, implementation of pubsub.EventTracer
func (pst *psTracer) Trace(evt *ps_pb.TraceEvent) {
	metricsPubsubTrace.WithLabelValues(evt.GetType().String()).Inc()
}
