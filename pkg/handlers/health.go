package handlers

import (
	"net/http"

	"github.com/netcracker/qubership-av-scan-service/pkg/clamav"
	"github.com/netcracker/qubership-av-scan-service/pkg/errors"
)

// HealthHandler handles health requests.
// It verifies that clamd is ready to be used.
type HealthHandler struct {
	clamd clamav.Clamd
}

func NewHealthHandler(clamd clamav.Clamd) *HealthHandler {
	return &HealthHandler{clamd: clamd}
}

func (h *HealthHandler) Handle(_ *http.Request) (any, error) {
	err := h.clamd.Ping()
	if err != nil {
		return nil, errors.ClamdPingError(err)
	}
	return nil, nil
}
