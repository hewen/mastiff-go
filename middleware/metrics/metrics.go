package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTPDuration records the duration of HTTP requests.
	HTTPDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// GRPCDuration records the duration of gRPC requests.
	GRPCDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "code"},
	)
)

func init() {
	prometheus.MustRegister(HTTPDuration)
	prometheus.MustRegister(GRPCDuration)
}
