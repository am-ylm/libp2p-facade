package streams

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricStreamOut = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_streams_out",
		Help: "Counts outgoing streams requests",
	}, []string{"protocol"})
	metricStreamOutDone = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_streams_out_done",
		Help: "Counts failed outgoing streams requests",
	}, []string{"protocol", "err"})
	metricStreamIn = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_streams_in",
		Help: "Counts incoming streams requests",
	}, []string{"protocol"})
	metricStreamInDone = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_streams_in_done",
		Help: "Counts failed outgoing streams requests",
	}, []string{"protocol", "err"})
)

func init() {
	_ = prometheus.Register(metricStreamOut)
	_ = prometheus.Register(metricStreamOutDone)
	_ = prometheus.Register(metricStreamIn)
	_ = prometheus.Register(metricStreamInDone)
}
