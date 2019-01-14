package handlers

import (
	"net/http"

	"github.com/emojify-app/api/logging"
)

// Health is a HTTP handler for serving health requests
type Health struct {
	logger logging.Logger
}

// NewHealth returns a new instance of the Health handler
func NewHealth(l logging.Logger) *Health {
	return &Health{l}
}

// ServeHTTP implements the http.Handler interface
func (h *Health) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	done := h.logger.HealthHandlerCalled()
	defer done(http.StatusOK, nil)

	rw.Write([]byte("OK"))
}
