package handlers

import "net/http"

// RequestHandler is an interface implemented by custom handlers.
// It is different from http.Handler to allow easier implementation of handlers.
// Custom handlers do not write response directly,
// instead they may return any value which could be marshaled to JSON, or error
type RequestHandler interface {
	Handle(r *http.Request) (any, error)
}
