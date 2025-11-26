package port

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MetricProxyRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_requests_total",
			Help: "Total number of HTTP requests processed by the proxy",
		},
		[]string{"code", "method"},
	)

	MetricProxyRequestsByHost = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxy_requests_by_host_total",
			Help: "Total number of HTTP requests processed by the proxy per host.",
		},
		[]string{"host", "method"},
	)

	MetricProxyRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "proxy_request_duration_seconds",
			Help:    "Latency of proxied requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	MetricProxyInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "proxy_inflight_requests",
			Help: "Current number of in-flight requests.",
		},
	)

	MetricProxyResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "proxy_response_size_bytes",
			Help:    "Size of responses.",
			Buckets: prometheus.ExponentialBuckets(512, 2, 10),
		},
		[]string{"method"},
	)
)
