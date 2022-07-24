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
	metricStreamOutFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_streams_out_failed",
		Help: "Counts failed outgoing streams requests",
	}, []string{"protocol", "err"})
	metricStreamOutSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_streams_out_success",
		Help: "Counts successful outgoing streams requests",
	}, []string{"protocol"})
	metricStreamOutActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "p2p_streams_out_active",
		Help: "Counts active outgoing streams requests",
	}, []string{"protocol"})
	metricStreamIn = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_streams_in",
		Help: "Counts incoming streams requests",
	}, []string{"protocol"})
	metricStreamInFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_streams_in_failed",
		Help: "Counts failed outgoing streams requests",
	}, []string{"protocol", "err"})
	metricStreamInSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "p2p_streams_in_success",
		Help: "Counts successful outgoing streams requests",
	}, []string{"protocol"})
	metricStreamInActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "p2p_streams_in_active",
		Help: "Counts active incoming streams requests",
	}, []string{"protocol"})
)

func init() {
	_ = prometheus.Register(metricStreamOut)
	_ = prometheus.Register(metricStreamOutFailed)
	_ = prometheus.Register(metricStreamOutSuccess)
	_ = prometheus.Register(metricStreamOutActive)
	_ = prometheus.Register(metricStreamIn)
	_ = prometheus.Register(metricStreamInFailed)
	_ = prometheus.Register(metricStreamInSuccess)
	_ = prometheus.Register(metricStreamInActive)
}
