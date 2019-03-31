package handlers

import (
	"net/http"

	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/emojify/protos/emojify"
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

}
