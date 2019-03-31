package handlers

import (
	"net/http"
	"net/http/httptest"

	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/stretchr/testify/mock"
)

func setupEmojiGetHandler() (*httptest.ResponseRecorder, *http.Request, *EmojifyGet) {
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

	h := NewEmojifyGet(logger, &mockEmojifyer)

	return rw, r, h
}
