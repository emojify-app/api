package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"google.golang.org/grpc/status"
)

// Health is a HTTP handler for serving health requests
type Health struct {
	logger logging.Logger
	em     emojify.Emojify
	cc     cache.CacheClient
}

// NewHealth returns a new instance of the Health handler
func NewHealth(l logging.Logger, em emojify.Emojify, cc cache.CacheClient) *Health {
	return &Health{l, em, cc}
}

// ServeHTTP implements the http.Handler interface
func (h *Health) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	done := h.logger.HealthHandlerCalled()
	defer done(http.StatusOK, nil)

	// check cache health
	resp, err := h.cc.Check(context.Background(), &cache.HealthCheckRequest{})
	if s := status.Convert(err); err != nil && s != nil {
		http.Error(rw, fmt.Sprintf("Error checking cache health %s", s.Message()), http.StatusInternalServerError)
		return
	}

	rw.Write([]byte("OK\n"))
	rw.Write([]byte(fmt.Sprintf("Cache status %d\n", resp.GetStatus())))
}
