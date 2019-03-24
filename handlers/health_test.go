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
)

func setupHealthTests() (*Health, *httptest.ResponseRecorder, *http.Request) {
	rw := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	l, _ := logging.New("test", "test", "localhost:8125", "error", "text")

	ec := &emojify.ClientMock{}
	ec.On("Check", mock.Anything, mock.Anything, mock.Anything).Return(&emojify.HealthCheckResponse{Status: emojify.HealthCheckResponse_SERVING}, nil)

	cc := &cache.ClientMock{}
	cc.On("Check", mock.Anything, mock.Anything, mock.Anything).Return(&cache.HealthCheckResponse{Status: cache.HealthCheckResponse_SERVING}, nil)

	return &Health{l, ec, cc}, rw, r
}

func TestHealthHandlerReturnsOK(t *testing.T) {
	h, rw, r := setupHealthTests()

	h.ServeHTTP(rw, r)

	assert.Contains(t, string(rw.Body.Bytes()), "OK")
}

func TestHealthHandlerReturns200(t *testing.T) {
	h, rw, r := setupHealthTests()

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
}
