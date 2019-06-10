package handlers

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"         // import image
	_ "image/png" // import image
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/golang/protobuf/ptypes/wrappers"
)

// EmojifyResponse is a Go representation of the protobuf QueryItem
// Status:
// UNKNOWN  = 0
// QUEUED   = 1
// FINISHED = 2
type EmojifyResponse struct {
	ID       string `json:"id"`
	Length   int32  `json:"length"`
	Position int32  `json:"position"`
	Status   string `json:"status"`
}

// FromQueryItem creates an EmojifyResponse from a emojify.QueryItem
func (er EmojifyResponse) FromQueryItem(qi *emojify.QueryItem) EmojifyResponse {
	return EmojifyResponse{
		ID:       qi.GetId(),
		Length:   qi.GetQueueLength(),
		Position: qi.GetQueuePosition(),
		Status:   qi.GetStatus().GetStatus().String(),
	}
}

// WriteJSON writes the response as json to a writer
func (er EmojifyResponse) WriteJSON(w io.Writer) {
	e := json.NewEncoder(w)
	e.Encode(&er)
}

// EmojifyPost is a http.Handler for Emojifying images
type EmojifyPost struct {
	logger  logging.Logger
	emojify emojify.EmojifyClient
}

// NewEmojifyPost returns a new instance of the Emojify handler
func NewEmojifyPost(l logging.Logger, e emojify.EmojifyClient) *EmojifyPost {
	return &EmojifyPost{l, e}
}

// ServeHTTP implements the handler function
func (e *EmojifyPost) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	done := e.logger.EmojifyHandlerPOSTCalled(r)

	// check the post body
	data, err := e.checkPostBody(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		done(http.StatusBadRequest, err)
		return
	}

	// validate the url
	var u *url.URL
	if u, err = e.validateURL(data); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		done(http.StatusBadRequest, nil)
		return
	}

	ecDone := e.logger.EmojifyHandlerCallCreate(u.String())
	resp, err := e.emojify.Create(context.Background(), &wrappers.StringValue{Value: u.String()})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		ecDone(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, err)
		return
	}

	// return the image key
	jr := EmojifyResponse{}.FromQueryItem(resp)
	rw.WriteHeader(http.StatusOK)
	jr.WriteJSON(rw)
	done(http.StatusOK, nil)
}

func (e *EmojifyPost) checkPostBody(r *http.Request) ([]byte, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e.logger.EmojifyHandlerNoPostBody()
		return nil, err
	}

	return data, nil
}

func (e *EmojifyPost) validateURL(data []byte) (*url.URL, error) {
	valid := govalidator.IsRequestURL(string(data))
	if valid == false {
		return nil, fmt.Errorf("%v is not a valid URL", string(data))
	}

	u, err := url.ParseRequestURI(string(data))
	if err != nil {
		e.logger.EmojifyHandlerInvalidURL(string(data), err)
		return nil, fmt.Errorf("unable to parse %v", string(data))
	}

	return u, nil
}

// hashFilename returns a md5 hash of the filename
func hashFilename(f string) string {
	h := md5.New()
	io.WriteString(h, f)

	return fmt.Sprintf("%x", h.Sum(nil))
}
