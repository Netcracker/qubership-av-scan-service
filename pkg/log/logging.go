package log

import (
	"context"
	"log/slog"
	"net/http"
)

// logKey is a hidden struct used as a unique context key
type logKey struct{}

// With sets given logger in the given request context,
// so that it could be retrieved in future using From (and only this way).
func With(r *http.Request, logger *slog.Logger) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), logKey{}, logger))
}

// From retrieves saved logger from given request context.
// If logger is not found, a default logger is returned.
func From(r *http.Request) *slog.Logger {
	v := r.Context().Value(logKey{})
	if logger, ok := v.(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}
