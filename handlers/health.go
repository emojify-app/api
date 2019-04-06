package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/protos/emojify"
	"google.golang.org/grpc/status"
)

// Health is a HTTP handler for serving health requests
type Health struct {
	logger logging.Logger
	ec     emojify.EmojifyClient
	cc     cache.CacheClient
}

// NewHealth returns a new instance of the Health handler
func NewHealth(l logging.Logger, ec emojify.EmojifyClient, cc cache.CacheClient) *Health {
	return &Health{l, ec, cc}
}

// ServeHTTP implements the http.Handler interface
func (h *Health) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	done := h.logger.HealthHandlerCalled()

	// check cache health
	resp, err := h.cc.Check(context.Background(), &cache.HealthCheckRequest{})
	if s := status.Convert(err); err != nil && s != nil {
		errString := fmt.Sprintf("Error checking cache health %s", s.Message())

		http.Error(rw, errString, http.StatusInternalServerError)
		done(http.StatusInternalServerError, fmt.Errorf(errString))
		return
	}

	// check emojify health
	respE, err := h.ec.Check(context.Background(), &emojify.HealthCheckRequest{})
	if s := status.Convert(err); err != nil && s != nil {
		errString := fmt.Sprintf("Error checking emojify health %s", s.Message())

		http.Error(rw, errString, http.StatusInternalServerError)
		done(http.StatusInternalServerError, err)
		return
	}

	rw.Write([]byte("OK\n"))
	rw.Write([]byte(fmt.Sprintf("Cache status %d\n", resp.GetStatus())))
	rw.Write([]byte(fmt.Sprintf("Emojify status %d\n", respE.GetStatus())))

	done(http.StatusOK, nil)
}
