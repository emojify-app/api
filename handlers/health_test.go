package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emojify-app/api/logging"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandlerReturnsOK(t *testing.T) {
	rw := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	l, _ := logging.New("test", "localhost:8125", "DEBUG", "text")

	h := Health{l}
	h.ServeHTTP(rw, r)

	assert.Equal(t, "OK", string(rw.Body.Bytes()))
}

func TestHealthHandlerReturns200(t *testing.T) {
	rw := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	l, _ := logging.New("test", "localhost:8125", "DEBUG", "text")

	h := Health{l}
	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
}
