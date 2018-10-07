package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/emojify-api/emojify"
	"github.com/stretchr/testify/assert"
)

var fileURL = "http://something.com/a.jpg"
var base64URL string

func setupCacheHandler() (*httptest.ResponseRecorder, *http.Request, *CacheHandler) {
	mockCache = emojify.MockCache{}
	logger := hclog.Default()
	base64URL = base64.StdEncoding.EncodeToString([]byte(fileURL))

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(
		"GET",
		"/cache?file="+base64URL,
		nil,
	)

	h := &CacheHandler{logger, &mockCache}

	return rw, r, h
}

func TestReturns405WhenNotGet(t *testing.T) {
	rw, _, h := setupCacheHandler()
	r := httptest.NewRequest("POST", "/", nil)

	h.Handle(rw, r)

	assert.Equal(t, http.StatusMethodNotAllowed, rw.Code)
}

func TestReturns400WhenInvalidFileParameter(t *testing.T) {
	rw, _, h := setupCacheHandler()
	r := httptest.NewRequest("GET", "/cache?file=something", nil)

	h.Handle(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
}

func TestReturns404WhenNoImageFoundInCache(t *testing.T) {
	rw, r, h := setupCacheHandler()
	mockCache.On("Get", fileURL).Return([]byte{}, fmt.Errorf("Not found"))
	h.Handle(rw, r)

	assert.Equal(t, http.StatusNotFound, rw.Code)
}

func TestReturns200WhenImageFound(t *testing.T) {
	rw, r, h := setupCacheHandler()
	mockCache.On("Get", fileURL).Return([]byte("abc"), nil)

	h.Handle(rw, r)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "abc", rw.Body.String())
}
