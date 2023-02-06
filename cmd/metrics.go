package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Latency = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "method_duration",
		Help: "RPC Duration",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	}, []string{"method"})
	ActiveSubscriptions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_subscriptions",
		Help: "The total number of active subscriptions",
	})
	FailedCalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "methods_error_count",
		Help: "Number of failed calls",
	}, []string{"method"})
	SucceedCalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "method_count",
		Help: "Number of successful calls",
	}, []string{"method"})
)

func init() {
	prometheus.Register(Latency)
	prometheus.Register(ActiveSubscriptions)
	prometheus.Register(FailedCalls)
	prometheus.Register(SucceedCalls)
}
