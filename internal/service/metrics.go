package service

import (
	"net/http"

	"github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy/internal/port"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMetrics() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}

func wrapProxyMetrics(inner http.Handler) http.Handler {
	h := inner

	h = promhttp.InstrumentHandlerCounter(
		port.MetricProxyRequestsTotal,
		h,
	)

	h = promhttp.InstrumentHandlerDuration(
		port.MetricProxyRequestDuration,
		h,
	)

	h = promhttp.InstrumentHandlerInFlight(
		port.MetricProxyInFlight,
		h,
	)

	h = promhttp.InstrumentHandlerResponseSize(
		port.MetricProxyResponseSize,
		h,
	)

	return h
}
