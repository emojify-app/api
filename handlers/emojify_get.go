package handlers

import (
	"context"
	"net/http"

	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/gorilla/mux"
)

// EmojifyGet is a http.Handler for querying the state of images
type EmojifyGet struct {
	logger  logging.Logger
	emojify emojify.EmojifyClient
}

// NewEmojifyGet returns a new instance of the Emojify handler
func NewEmojifyGet(l logging.Logger, e emojify.EmojifyClient) *EmojifyGet {
	return &EmojifyGet{l, e}
}

// ServeHTTP implements the handler function
func (e *EmojifyGet) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	done := e.logger.EmojifyHandlerGETCalled(r)

	vars := mux.Vars(r) // Get varaibles from the request path

	// check the parameters contains a valid url
	id := vars["id"]
	if id == "" {
		done(http.StatusBadRequest, nil)

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// errors from the emojify api should be treated like 404, could just be a
	// queue or cache item missing
	qDone := e.logger.EmojifyHandlerCallQuery(id)
	qi, err := e.emojify.Query(context.Background(), &wrappers.StringValue{Value: id})
	if err != nil {
		qDone(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, err)

		rw.WriteHeader(http.StatusNotFound)
		return
	}

	qDone(http.StatusOK, nil)

	EmojifyResponse{}.FromQueryItem(qi).WriteJSON(rw)
	done(http.StatusOK, nil)
}
