package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var mockEmojifyer emojify.ClientMock

func resetEmojifyMock() {
	mockEmojifyer.ExpectedCalls = make([]*mock.Call, 0)
}

func setupEmojiPostHandler() (*httptest.ResponseRecorder, *http.Request, *EmojifyPost) {
	mockEmojifyer = emojify.ClientMock{}

	mockEmojifyer.On(
		"Create",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&emojify.QueryItem{
			Id:            "abc",
			QueuePosition: 2,
			QueueLength:   4,
			Status:        &emojify.QueryStatus{Status: emojify.QueryStatus_QUEUED},
		},
		nil,
	)

	logger, _ := logging.New("test", "test", "localhost:8125", "error", "text")

	rw := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)

	h := NewEmojifyPost(logger, &mockEmojifyer)

	return rw, r, h
}

func TestReturnsBadRequestIfBodyLessThan8(t *testing.T) {
	rw, r, h := setupEmojiPostHandler()

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, " is not a valid URL\n", string(rw.Body.Bytes()))
}

func TestReturnsInvalidURLIfBodyNotURL(t *testing.T) {
	rw, r, h := setupEmojiPostHandler()
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("httsddfdfdf/cc")))

	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "httsddfdfdf/cc is not a valid URL\n", string(rw.Body.Bytes()))
}

func TestCallsEmojifyAndOK(t *testing.T) {
	rw, r, h := setupEmojiPostHandler()

	u, _ := url.Parse(fileURL)
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(u.String())))
	h.ServeHTTP(rw, r)

	qi := EmojifyResponse{}
	json.Unmarshal(rw.Body.Bytes(), &qi)

	assert.Equal(t, "abc", qi.ID)
	assert.Equal(t, int32(2), qi.Position)
	assert.Equal(t, int32(4), qi.Length)
}

func TestCallsEmojifyAndNotOK(t *testing.T) {
	rw, r, h := setupEmojiPostHandler()
	resetEmojifyMock()

	mockEmojifyer.On(
		"Create",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		nil,
		grpc.Errorf(codes.Internal, "Boom"),
	)

	u, _ := url.Parse(fileURL)
	r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(u.String())))
	h.ServeHTTP(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}
