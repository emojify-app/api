package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupHealthTests(ce, ee error) (*Health, *httptest.ResponseRecorder, *http.Request) {
	rw := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	l, _ := logging.New("test", "test", "localhost:8125", "error", "text")

	cc := &cache.ClientMock{}
	cc.On("Check", mock.Anything, mock.Anything, mock.Anything).Return(&cache.HealthCheckResponse{Status: cache.HealthCheckResponse_SERVING}, ce)

	ec := &emojify.ClientMock{}
	ec.On("Check", mock.Anything, mock.Anything, mock.Anything).Return(&emojify.HealthCheckResponse{Status: emojify.HealthCheckResponse_SERVING}, ee)

	return &Health{l, ec, cc}, rw, r
}

func TestHealthHandlerReturns200(t *testing.T) {
	h, rw, r := setupHealthTests(nil, nil)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestHealthHandlerReturns500OnCacheError(t *testing.T) {
	h, rw, r := setupHealthTests(status.Error(codes.Internal, "boom"), nil)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func TestHealthHandlerReturns500OnEmojifyError(t *testing.T) {
	h, rw, r := setupHealthTests(nil, status.Error(codes.Internal, "boom"))

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}
