package pubsub

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricPubsubListening = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "p2p:pubsub:listen",
		Help: "Counts topics that we listen to",
	}, []string{"topic"})
	metricPubsubOut = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p:pubsub:out",
		Help: "Counts outgoing pubsub messages",
	}, []string{"topic"})
	metricPubsubIn = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p:pubsub:in",
		Help: "Counts incoming pubsub messages",
	}, []string{"topic"})
	metricPubsubInDropped = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p:pubsub:in:dropped",
		Help: "Counts incoming pubsub messages that were dropped",
	}, []string{"topic"})
)

func init() {
	_ = prometheus.Register(metricPubsubListening)
	_ = prometheus.Register(metricPubsubOut)
	_ = prometheus.Register(metricPubsubIn)
	_ = prometheus.Register(metricPubsubInDropped)
}
