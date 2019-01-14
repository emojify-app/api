package handlers

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/logging"
	"github.com/stretchr/testify/assert"
)

var fileURL = "http://something.com/a.jpg"
var base64URL string

func setupCacheHandler() (*httptest.ResponseRecorder, *http.Request, *Cache) {
	mockCache = emojify.MockCache{}
	base64URL = base64.StdEncoding.EncodeToString([]byte(fileURL))
	logger, _ := logging.New("test", "localhost:8125", "DEBUG")

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(
		"GET",
		"/cache?file="+base64URL,
		nil,
	)

	h := &Cache{logger, &mockCache}

	return rw, r, h
}

func TestReturns400WhenInvalidFileParameter(t *testing.T) {
	rw, _, h := setupCacheHandler()
	r := httptest.NewRequest("GET", "/cache", nil)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
}

func TestReturns404WhenNoImageFoundInCache(t *testing.T) {
	rw, r, h := setupCacheHandler()
	mockCache.On("Get", base64URL).Return([]byte{}, nil)
	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusNotFound, rw.Code)
}

func TestReturns200WhenImageFound(t *testing.T) {
	rw, r, h := setupCacheHandler()
	mockCache.On("Get", base64URL).Return([]byte("abc"), nil)

	h.ServeHTTP(rw, r)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "abc", rw.Body.String())
}
