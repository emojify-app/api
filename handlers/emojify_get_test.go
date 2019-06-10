package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupEmojiGetHandler(id string) (*httptest.ResponseRecorder, *http.Request, *EmojifyGet) {
	mockEmojifyer = emojify.ClientMock{}

	mockEmojifyer.On(
		"Query",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		&emojify.QueryItem{
			Id:            id,
			QueuePosition: 2,
			QueueLength:   4,
			Status:        &emojify.QueryStatus{Status: emojify.QueryStatus_QUEUED},
		},
		nil,
	)

	logger, _ := logging.New("test", "test", "localhost:8125", "error", "text")

	// Set the gorilla mux vars for testing
	r := httptest.NewRequest("GET", "/", nil)
	if id != "" {
		r = mux.SetURLVars(r, map[string]string{"id": "abc123"})
	}
	rw := httptest.NewRecorder()

	h := NewEmojifyGet(logger, &mockEmojifyer)

	return rw, r, h
}

func TestGetReturnsBadRequestWhenNoId(t *testing.T) {
	rr, r, e := setupEmojiGetHandler("")

	e.ServeHTTP(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetReturns404ErrorWhenCacheError(t *testing.T) {
	rr, r, e := setupEmojiGetHandler("abc123")
	resetEmojifyMock()
	mockEmojifyer.On(
		"Query",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(
		nil,
		fmt.Errorf("boom"),
	)

	e.ServeHTTP(rr, r)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetReturnsQueueItemWhenOk(t *testing.T) {
	rr, r, e := setupEmojiGetHandler("abc123")

	e.ServeHTTP(rr, r)

	er := EmojifyResponse{}
	json.Unmarshal(rr.Body.Bytes(), &er)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, int32(2), er.Position)
}
