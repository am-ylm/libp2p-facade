package p2pfacade

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "p2p_peers_connected",
		Help: "Count connected peers",
	}, []string{"pid"})
)

func init() {
	_ = prometheus.Register(metricConnections)
}
