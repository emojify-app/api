package handlers

import (
	"fmt"
	"net/http"

	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/logging"
)

// Health is a HTTP handler for serving health requests
type Health struct {
	logger logging.Logger
	em     emojify.Emojify
}

// NewHealth returns a new instance of the Health handler
func NewHealth(l logging.Logger, em emojify.Emojify) *Health {
	return &Health{l, em}
}

// ServeHTTP implements the http.Handler interface
func (h *Health) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	done := h.logger.HealthHandlerCalled()
	defer done(http.StatusOK, nil)

	// check facebox health
	info, err := h.em.Health()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write([]byte("OK\n"))
	rw.Write([]byte(fmt.Sprintf("Facebox version: %d\n", info.Version)))
}
