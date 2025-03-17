package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/netcracker/qubership-av-scan-service/pkg/errors"
	"github.com/netcracker/qubership-av-scan-service/pkg/handlers"
	"github.com/netcracker/qubership-av-scan-service/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/xid"
)

// requestHandlerAdapter adapts handlers.RequestHandler to http.Handler interface.
// If error is returned, it is logged and written in response.
// If there is no error and result is present, adapter tries to write result in response.
func requestHandlerAdapter(reqHandler handlers.RequestHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res, err := reqHandler.Handle(r)

		logger := log.From(r)
		if err != nil {
			handleError(w, logger, err)
			return
		}

		if res != nil {
			err = writeResponse(w, res)
			if err != nil {
				handleError(w, logger, err)
			}
		}
	})
}

// loggingMiddleware logs high-level information about request start/end.
// It also saves logger in the request context with additional fields for future use
func loggingMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		logger := logger.With("reqId", xid.New(), "method", req.Method, "url", req.URL)
		logger.Debug("request received")

		// defer logging of outgoing response
		lrw := newLoggingResponseWriter(w)
		defer func() {
			if lrw.statusCode > 400 {
				logger.Error("request completed with error",
					"status", lrw.statusCode,
					"duration", time.Since(start))
			} else {
				logger.Info("request completed successfully",
					"status", lrw.statusCode,
					"duration", time.Since(start))
			}
		}()

		next.ServeHTTP(lrw, log.With(req, logger))
	})
}

// panicRecoveryMiddleware recovers from panic by logging the error and
// writing it in response. Should be used as close as possible to actual handler
// so that panic do not unwind too much other handlers (like metrics/logging).
func panicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			r := recover()
			if r != nil {
				var err error
				switch t := r.(type) {
				case error:
					err = t
				default:
					err = fmt.Errorf("%v", r)
				}
				logger := log.From(request)
				logger.Error("handler panicked", "stacktrace", debug.Stack())
				handleError(writer, logger, err)
			}
		}()
		next.ServeHTTP(writer, request)
	})
}

// metricsMiddleware collects requests total, duration, size and inflight count
// metrics for given HTTP handler
func metricsMiddleware(
	next http.Handler,
	registry *prometheus.Registry,
	name string,
) http.Handler {
	reg := prometheus.WrapRegistererWith(prometheus.Labels{"handler": name}, registry)
	requests := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total"},
		[]string{"code"},
	)
	duration := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{Name: "http_request_duration_seconds"},
		[]string{"code"},
	)
	size := promauto.With(reg).NewSummaryVec(
		prometheus.SummaryOpts{Name: "http_request_size_bytes"},
		[]string{"code"},
	)
	inflight := promauto.With(reg).NewGauge(
		prometheus.GaugeOpts{Name: "http_requests_inflight"},
	)

	next = promhttp.InstrumentHandlerCounter(requests, next)
	next = promhttp.InstrumentHandlerDuration(duration, next)
	next = promhttp.InstrumentHandlerRequestSize(size, next)
	next = promhttp.InstrumentHandlerInFlight(inflight, next)
	return next
}

// loggingResponseWriter is wrapper around http.ResponseWriter which
// saves statusCode for future use in logging middleware
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// writeResponse marshals given value as JSON and writes it in response body.
// If any error happens during write, it is returned as is.
func writeResponse(resp http.ResponseWriter, v any) error {
	if v != nil {
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}

		resp.Header().Add("Content-Type", "application/json")
		_, err = fmt.Fprintf(resp, "%s\n", data)
		if err != nil {
			return err
		}
	}
	return nil
}

// handleError logs given error and tries to write it in response
func handleError(w http.ResponseWriter, l *slog.Logger, err error) {
	if e, ok := err.(*errors.APIError); ok {
		l.Error(
			"request error",
			"reason", e.Reason,
			"details", e.Details,
			"code", e.Code,
			"status", e.Status,
		)
		b, err := json.MarshalIndent(e, "", "  ")
		if err != nil {
			l.Error("failed to marshall response error", "error", err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(e.Status)
		_, err = fmt.Fprintf(w, "%s\n", b)
		if err != nil {
			l.Error("failed to handleError response error", "error", err)
			return
		}
	} else {
		handleError(w, l, errors.UnexpectedError(err))
		return
	}
	return
}
