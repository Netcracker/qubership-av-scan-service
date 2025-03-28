package router

import (
	"log/slog"
	"net/http"

	"github.com/netcracker/qubership-av-scan-service/pkg/clamav"
	"github.com/netcracker/qubership-av-scan-service/pkg/handlers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter returns a http.Handler which handles all application endpoints.
// Under the hood, it uses dedicated handlers and special middleware
// for metrics, logging, panic recovery, etc
func NewRouter(clamd clamav.Clamd, logger *slog.Logger) http.Handler {
	if clamd == nil {
		panic("Server MUST be provided with ClamD instance")
	}
	if logger == nil {
		logger = slog.Default()
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	registry.MustRegister(clamav.NewMetricsCollector(clamd, logger))

	m := http.NewServeMux()
	m.Handle("POST /api/v1/scan", newScanHandler(clamd, registry))
	m.Handle("GET /health", newHealthHandler(clamd, registry))
	m.Handle("GET /metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	return loggingMiddleware(m, logger)
}

func newHealthHandler(clamd clamav.Clamd, registry *prometheus.Registry) http.Handler {
	handler := requestHandlerAdapter(handlers.NewHealthHandler(clamd))
	return metricsMiddleware(panicRecoveryMiddleware(handler), registry, "health")
}

func newScanHandler(clamd clamav.Clamd, registry *prometheus.Registry) http.Handler {
	handler := requestHandlerAdapter(handlers.NewScanHandler(clamd, registry))
	return metricsMiddleware(panicRecoveryMiddleware(handler), registry, "scan")
}
