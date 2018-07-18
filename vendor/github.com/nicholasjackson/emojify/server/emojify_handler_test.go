package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/machinebox/sdk-go/facebox"
	"github.com/nicholasjackson/emojify/emojify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var mockFetcher emojify.MockFetcher
var mockEmojifyer emojify.MockEmojify

func setupEmojiHandler() (*httptest.ResponseRecorder, *http.Request, *emojiHandler) {
	mockFetcher = emojify.MockFetcher{}
	mockEmojifyer = emojify.MockEmojify{}

	rw := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)

	h := &emojiHandler{&mockEmojifyer, &mockFetcher}

	return rw, r, h
}

func TestReturnsNotAllowedIfMethodNotPOST(t *testing.T) {
	rw := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	handler := emojiHandler{}
	handler.Handle(rw, r)

	assert.Equal(t, http.StatusMethodNotAllowed, rw.Code)
}

func TestReturnsBadRequestIfBodyLessThan8(t *testing.T) {
	rw, r, h := setupEmojiHandler()

	h.Handle(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, " is not a valid URL", string(rw.Body.Bytes()))
}

func TestReturnsInvalidURLIfBodyNotURL(t *testing.T) {
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("httsddfdfdf/cc")))

	h.Handle(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "httsddfdfdf/cc is not a valid URL", string(rw.Body.Bytes()))
}

func TestReturnsInternalServerErrorWhenCantFetchImage(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))
	mockFetcher.On("FetchImage", url).Return(nil, fmt.Errorf("Unable to get image"))

	h.Handle(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func TestRetrunsInternalServerErrorWhenDataNotImage(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))
	mockFetcher.On("FetchImage", url).Return(bytes.NewReader([]byte("")), nil)
	mockFetcher.On("ReaderToImage", mock.Anything).Return(nil, fmt.Errorf("Invalid image"))

	h.Handle(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Equal(t, "invalid image format "+url, string(rw.Body.Bytes()))
}

func TestRetrunsInternalServerErrorWhenDataNoFaces(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))
	mockFetcher.On("FetchImage", url).Return(bytes.NewReader([]byte("")), nil)
	mockFetcher.On("ReaderToImage", mock.Anything).Return(image.NewUniform(color.Black), nil)
	mockEmojifyer.On("GetFaces", mock.Anything).Return(nil, fmt.Errorf("No faces"))

	h.Handle(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Equal(t, "unable to find faces in "+url, string(rw.Body.Bytes()))
}

func TestReturnsInternalServerErrorWhenUnableToEmojify(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))
	mockFetcher.On("FetchImage", url).Return(bytes.NewReader([]byte("")), nil)
	mockFetcher.On("ReaderToImage", mock.Anything).Return(image.NewUniform(color.Black), nil)
	mockEmojifyer.On("GetFaces", mock.Anything).Return(make([]facebox.Face, 0), nil)
	mockEmojifyer.On("Emojimise", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Cant emojify"))

	h.Handle(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Equal(t, "unable to emojify "+url, string(rw.Body.Bytes()))
}

func TestSetsHeaderWhenOK(t *testing.T) {
	url := "https://something.com"
	rw, r, h := setupEmojiHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(url)))
	mockFetcher.On("FetchImage", url).Return(bytes.NewReader([]byte("")), nil)
	mockFetcher.On("ReaderToImage", mock.Anything).Return(image.NewUniform(color.Black), nil)
	mockEmojifyer.On("GetFaces", mock.Anything).Return(make([]facebox.Face, 0), nil)
	mockEmojifyer.On("Emojimise", mock.Anything, mock.Anything).Return(image.NewRGBA64(image.Rect(0, 0, 0, 0)), nil)

	h.Handle(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, rw.HeaderMap.Get("content-type"), "image/png")
}
