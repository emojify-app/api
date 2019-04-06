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
	resp, errC := h.cc.Check(context.Background(), &cache.HealthCheckRequest{})
	if s := status.Convert(errC); errC != nil && s != nil {
		errString := fmt.Sprintf("Error checking cache health %s", s.Message())

		h.logger.Log().Error("Health handler error", "error", errString)
		rw.Write([]byte(fmt.Sprintf("Cache status: %s\n", errString)))
	} else {
		rw.Write([]byte(fmt.Sprintf("Cache status: %d\n", resp.GetStatus())))
	}

	// check emojify health
	respE, errE := h.ec.Check(context.Background(), &emojify.HealthCheckRequest{})
	if s := status.Convert(errE); errE != nil && s != nil {
		errString := fmt.Sprintf("Error checking emojify health %s", s.Message())

		h.logger.Log().Error("Health handler error", "error", errString)
		rw.Write([]byte(fmt.Sprintf("Emojify status: %s\n", errString)))
	} else {
		rw.Write([]byte(fmt.Sprintf("Emojify status: %d\n", respE.GetStatus())))
	}

	if errC != nil || errE != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		done(http.StatusInternalServerError, nil)
		return
	}

	done(http.StatusOK, nil)
}
