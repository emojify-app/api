package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
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

	// check facebox health
	info, err := h.em.Health()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// check cache health
	_, err = h.cc.Get(context.Background(), &wrappers.StringValue{Value: "notexist"})
	if s := status.Convert(err); s != nil && s.Code() != codes.NotFound {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write([]byte("OK\n"))
	rw.Write([]byte(fmt.Sprintf("Facebox version: %d\n", info.Version)))
}
