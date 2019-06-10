package handlers

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var fileURL = "http://something.com/a.jpg"
var base64URL string
var mockCache cache.ClientMock

func setupCacheHandler() (*httptest.ResponseRecorder, *http.Request, *Cache) {
	mockCache = cache.ClientMock{}
	base64URL = base64.StdEncoding.EncodeToString([]byte(fileURL))
	logger, _ := logging.New("test", "test", "localhost:8125", "error", "text")

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(
		"GET",
		"/",
		nil,
	)

	// Set the gorilla mux vars for testing
	r = mux.SetURLVars(r, map[string]string{"id": base64URL})

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
	mockCache.On(
		"Get",
		mock.Anything,
		&wrappers.StringValue{Value: base64URL},
		mock.Anything,
	).Return(nil, status.Error(codes.NotFound, "Not found"))

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusNotFound, rw.Code)
}

func TestReturns200WhenImageFound(t *testing.T) {
	rw, r, h := setupCacheHandler()
	mockCache.On(
		"Get",
		mock.Anything,
		&wrappers.StringValue{Value: base64URL},
		mock.Anything,
	).Return(&cache.CacheItem{Data: []byte("abc")}, nil)

	h.ServeHTTP(rw, r)

	assert.Equal(t, 200, rw.Code)
	assert.Equal(t, "abc", rw.Body.String())
}
