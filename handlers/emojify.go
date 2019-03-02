package handlers

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	_ "image/jpeg" // import image
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/golang/protobuf/ptypes/wrappers"
)

// Emojify is a http.Handler for Emojifying images
type Emojify struct {
	emojifyer emojify.Emojify
	fetcher   emojify.Fetcher
	logger    logging.Logger
	cache     cache.CacheClient
}

// NewEmojify returns a new instance of the Emojify handler
func NewEmojify(e emojify.Emojify, f emojify.Fetcher, l logging.Logger, c cache.CacheClient) *Emojify {
	return &Emojify{e, f, l, c}
}

// ServeHTTP implements the handler function
func (e *Emojify) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	done := e.logger.EmojifyHandlerCalled(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e.logger.EmojifyHandlerNoPostBody()

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	var u *url.URL

	if u, err = validateURL(data); err != nil {
		e.logger.EmojifyHandlerInvalidURL(string(data), err)
		done(http.StatusBadRequest, nil)

		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	key := hashFilename(u.String())

	ccDone := e.logger.EmojifyHandlerCacheCheck(key)
	ok, err := e.cache.Exists(context.Background(), &wrappers.StringValue{Value: key})
	if err != nil {
		ccDone(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, nil)

		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if ok.Value {
		// cache found message
		ccDone(http.StatusOK, nil)
		done(http.StatusOK, nil)

		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(key))
		return
	}

	// log cache file not found
	ccDone(http.StatusNotFound, nil)

	fiDone := e.logger.EmojifyHandlerFetchImage(u.String())
	f, err := e.fetcher.FetchImage(u.String())
	if err != nil {
		fiDone(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, nil)

		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	fiDone(http.StatusOK, nil)

	img, err := e.fetcher.ReaderToImage(f)
	if err != nil {
		e.logger.EmojifyHandlerInvalidImage(u.String(), err)
		done(http.StatusInternalServerError, nil)

		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	ffDone := e.logger.EmojifyHandlerFindFaces(u.String())
	doneChan := make(chan emojify.FindFaceResponse)
	e.emojifyer.GetFaces(f, doneChan)
	resp := <-doneChan

	if resp.Error != nil {
		ffDone(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, nil)
		http.Error(rw, resp.Error.Error(), http.StatusInternalServerError)
		return
	}
	ffDone(http.StatusOK, nil)

	emDone := e.logger.EmojifyHandlerEmojify(u.String())
	i, err := e.emojifyer.Emojimise(img, resp.Faces)
	if err != nil {
		emDone(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, nil)

		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	emDone(http.StatusOK, nil)

	// save the image
	out := new(bytes.Buffer)
	err = png.Encode(out, i)
	if err != nil {
		e.logger.EmojifyHandlerImageEncodeError(u.String(), err)

		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// save the cache
	cpDone := e.logger.EmojifyHandlerCachePut(u.String())
	_, err = e.cache.Put(context.Background(), &cache.CacheItem{Id: key, Data: out.Bytes()})
	if err != nil {
		cpDone(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, nil)

		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	cpDone(http.StatusOK, nil)

	rw.Write([]byte(key))
	done(http.StatusOK, nil)
}

func validateURL(data []byte) (*url.URL, error) {
	valid := govalidator.IsRequestURL(string(data))
	if valid == false {
		return nil, fmt.Errorf("%v is not a valid URL", string(data))
	}

	u, err := url.ParseRequestURI(string(data))
	if err != nil {
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
