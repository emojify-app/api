package handlers

import (
	"bytes"
	"fmt"
	_ "image/jpeg" // import image
	"image/png"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/logging"
)

// Emojify is a http.Handler for Emojifying images
type Emojify struct {
	emojifyer emojify.Emojify
	fetcher   emojify.Fetcher
	logger    logging.Logger
	cache     emojify.Cache
}

// NewEmojify returns a new instance of the Emojify handler
func NewEmojify(e emojify.Emojify, f emojify.Fetcher, l logging.Logger, c emojify.Cache) *Emojify {
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
		e.logger.EmojifyHandlerInvalidURL(u.String(), err)
		done(http.StatusBadRequest, nil)

		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	key := emojify.HashFilename(u.String())

	ccDone := e.logger.EmojifyHandlerCacheCheck(key)
	ok, err := e.cache.Exists(key)
	if err != nil {
		ccDone(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, nil)

		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if ok {
		// cache found message
		ccDone(http.StatusOK, nil)
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
	faces, err := e.emojifyer.GetFaces(f)
	if err != nil {
		ffDone(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, nil)

		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	ffDone(http.StatusOK, nil)

	emDone := e.logger.EmojifyHandlerEmojify(u.String())
	i, err := e.emojifyer.Emojimise(img, faces)
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
	err = e.cache.Put(key, out.Bytes())
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
