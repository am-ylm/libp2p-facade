package pubsub

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricPubsubListening = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "p2p_pubsub_listen",
		Help: "Counts topics that we listen to",
	}, []string{"topic"})
	metricPubsubOut = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_out",
		Help: "Counts outgoing pubsub messages",
	}, []string{"topic"})
	metricPubsubIn = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_in",
		Help: "Counts incoming pubsub messages",
	}, []string{"topic"})
	metricPubsubInDropped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_in_dropped",
		Help: "Counts incoming pubsub messages that were dropped",
	}, []string{"topic"})
	metricsPubsubTrace = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_pubsub_trace",
		Help: "Tracks pubsub tracing events",
	}, []string{"tp"})
)

func init() {
	_ = prometheus.Register(metricPubsubListening)
	_ = prometheus.Register(metricPubsubOut)
	_ = prometheus.Register(metricPubsubIn)
	_ = prometheus.Register(metricPubsubInDropped)
}
