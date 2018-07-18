package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthHandlerReturnsOK(t *testing.T) {
	rw := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)

	handler := healthHandler{}
	handler.Handle(rw, r)

	assert.Equal(t, "OK", string(rw.Body.Bytes()))
}

func TestHealthHandlerReturns200(t *testing.T) {
	rw := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)

	handler := healthHandler{}
	handler.Handle(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
}
