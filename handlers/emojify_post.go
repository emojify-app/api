package handlers

import (
	"context"
	"crypto/md5"
	"fmt"         // import image
	_ "image/png" // import image
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/golang/protobuf/ptypes/wrappers"
)

// EmojifyPost is a http.Handler for Emojifying images
type EmojifyPost struct {
	logger  logging.Logger
	emojify emojify.EmojifyClient
	cache   cache.CacheClient
}

// NewEmojifyPost returns a new instance of the Emojify handler
func NewEmojifyPost(l logging.Logger, e emojify.EmojifyClient, c cache.CacheClient) *EmojifyPost {
	return &EmojifyPost{l, e, c}
}

// ServeHTTP implements the handler function
func (e *EmojifyPost) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	done := e.logger.EmojifyHandlerCalled(r)

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

	// check the cache
	key := hashFilename(u.String())
	ok, err := e.checkCache(key)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		done(http.StatusInternalServerError, nil)
		return
	}

	// cache found message if ok
	if ok {
		rw.WriteHeader(http.StatusNotModified)
		rw.Write([]byte(key))
		done(http.StatusNotModified, nil)
		return
	}

	// return the image key
	rw.WriteHeader(http.StatusTeapot)
	//rw.Write([]byte(key))
	//done(http.StatusOK, nil)
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

func (e *EmojifyPost) checkCache(key string) (bool, error) {
	ccDone := e.logger.EmojifyHandlerCacheCheck(key)
	ok, err := e.cache.Exists(context.Background(), &wrappers.StringValue{Value: key})

	if err != nil {
		ccDone(http.StatusInternalServerError, err)
		return false, err
	}

	// log cache file not found
	if !ok.Value {
		ccDone(http.StatusNotFound, nil)
		return false, nil
	}

	ccDone(http.StatusOK, nil)
	return true, err
}

// hashFilename returns a md5 hash of the filename
func hashFilename(f string) string {
	h := md5.New()
	io.WriteString(h, f)

	return fmt.Sprintf("%x", h.Sum(nil))
}
