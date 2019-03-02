package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/machinebox/sdk-go/facebox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var mockFetcher emojify.MockFetcher
var mockEmojifyer emojify.MockEmojify
var mockCache cache.ClientMock

func setupEmojiHandler() (*httptest.ResponseRecorder, *http.Request, *Emojify) {
	mockFetcher = emojify.MockFetcher{}
	mockEmojifyer = emojify.MockEmojify{}
	mockCache = cache.ClientMock{}
	logger, _ := logging.New("test", "test", "localhost:8125", "error", "text")

	rw := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)

	h := NewEmojify(&mockEmojifyer, &mockFetcher, logger, &mockCache)

	return rw, r, h
}

func TestReturnsBadRequestIfBodyLessThan8(t *testing.T) {
	rw, r, h := setupEmojiHandler()

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, " is not a valid URL\n", string(rw.Body.Bytes()))
}

func TestReturnsInvalidURLIfBodyNotURL(t *testing.T) {
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("httsddfdfdf/cc")))

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "httsddfdfdf/cc is not a valid URL\n", string(rw.Body.Bytes()))
}

func TestReturns200IfImageIsCached(t *testing.T) {
	u, _ := url.Parse(fileURL)

	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(u.String())))
	mockCache.On(
		"Exists",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&wrappers.BoolValue{Value: true},
		nil,
	)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
}

func TestReturnsInternalServerErrorWhenCacheError(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))

	mockFetcher.On(
		"FetchImage",
		url,
	).Return(
		nil,
		fmt.Errorf("Unable to get image"),
	)

	mockCache.On(
		"Exists",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		nil,
		status.Error(codes.Internal, ""),
	)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func TestReturnsInternalServerErrorWhenCantFetchImage(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))

	mockFetcher.On(
		"FetchImage",
		url,
	).Return(
		nil,
		fmt.Errorf("Unable to get image"),
	)

	mockCache.On(
		"Exists",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&wrappers.BoolValue{Value: false},
		nil,
	)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func TestReturnsInternalServerErrorWhenDataNotImage(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))

	mockFetcher.On(
		"FetchImage",
		url,
	).Return(
		bytes.NewReader([]byte("")),
		nil,
	)

	mockFetcher.On(
		"ReaderToImage",
		mock.Anything,
	).Return(
		nil,
		fmt.Errorf("Invalid image"),
	)

	mockCache.On(
		"Exists",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&wrappers.BoolValue{Value: false},
		nil,
	)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Equal(t, "Invalid image\n", string(rw.Body.Bytes()))
}

func TestReturnsInternalServerErrorWhenDataNoFaces(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))

	mockFetcher.On(
		"FetchImage",
		url,
	).Return(
		bytes.NewReader([]byte("")),
		nil,
	)

	mockFetcher.On(
		"ReaderToImage",
		mock.Anything,
	).Return(
		image.NewUniform(color.Black),
		nil,
	)

	mockEmojifyer.On(
		"GetFaces",
		mock.Anything,
		mock.Anything,
	).Return(
		nil,
		fmt.Errorf("No faces"),
	)

	mockCache.On(
		"Exists",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&wrappers.BoolValue{Value: false},
		nil,
	)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Equal(t, "No faces\n", string(rw.Body.Bytes()))
}

func TestReturnsInternalServerErrorWhenUnableToEmojify(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))

	mockFetcher.On(
		"FetchImage",
		url,
	).Return(
		bytes.NewReader([]byte("")),
		nil,
	)

	mockFetcher.On(
		"ReaderToImage",
		mock.Anything,
	).Return(
		image.NewUniform(color.Black),
		nil,
	)

	mockEmojifyer.On(
		"GetFaces",
		mock.Anything,
		mock.Anything,
	).Return(
		make([]facebox.Face, 0),
		nil,
	)

	mockEmojifyer.On(
		"Emojimise",
		mock.Anything,
		mock.Anything,
	).Return(
		nil,
		fmt.Errorf("Cant emojify"),
	)

	mockCache.On(
		"Exists",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&wrappers.BoolValue{Value: false},
		nil,
	)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Equal(t, "Cant emojify\n", string(rw.Body.Bytes()))
}

func TestReturnsInternalServiceErrorWhenUnableToSaveCache(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))

	mockFetcher.On(
		"FetchImage",
		url,
	).Return(
		bytes.NewReader([]byte("")),
		nil,
	)

	mockFetcher.On(
		"ReaderToImage",
		mock.Anything,
	).Return(
		image.NewUniform(color.Black),
		nil,
	)

	mockEmojifyer.On(
		"GetFaces",
		mock.Anything,
		mock.Anything,
	).Return(
		make([]facebox.Face, 0),
		nil,
	)

	mockEmojifyer.On(
		"Emojimise",
		mock.Anything,
		mock.Anything,
	).Return(
		image.NewRGBA64(
			image.Rect(0, 0, 0, 0)),
		nil,
	)

	mockCache.On(
		"Put",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		status.Error(codes.Internal, ""),
	)

	mockCache.On(
		"Exists",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&wrappers.BoolValue{Value: false},
		nil,
	)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func TestReturnsStatusOKOnSuccess(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))
	img := image.NewRGBA64(image.Rect(0, 0, 400, 400))

	mockFetcher.On(
		"FetchImage",
		url,
	).Return(
		bytes.NewReader([]byte("")),
		nil,
	)

	mockFetcher.On(
		"ReaderToImage",
		mock.Anything,
	).Return(
		image.NewUniform(color.Black),
		nil,
	)

	mockEmojifyer.On(
		"GetFaces",
		mock.Anything,
		mock.Anything,
	).Return(
		make([]facebox.Face, 0),
		nil,
	)

	mockEmojifyer.On(
		"Emojimise",
		mock.Anything,
		mock.Anything,
	).Return(
		img,
		nil,
	)

	mockCache.On(
		"Put",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&wrappers.StringValue{Value: "abc"},
		nil,
	)

	mockCache.On(
		"Exists",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&wrappers.BoolValue{Value: false},
		nil,
	)

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
}
