package handlers

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"image"
	"image/jpeg"  // import image
	_ "image/png" // import image
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/machinebox/sdk-go/facebox"
)

// Emojify is a http.Handler for Emojifying images
type Emojify struct {
	emojifyer      emojify.Emojify
	fetcher        emojify.Fetcher
	logger         logging.Logger
	cache          cache.CacheClient
	faceboxWorkers int32
	activeWorkers  int32
	workerTimeout  time.Duration
}

// NewEmojify returns a new instance of the Emojify handler
func NewEmojify(e emojify.Emojify, f emojify.Fetcher, l logging.Logger, c cache.CacheClient) *Emojify {
	return &Emojify{e, f, l, c, 1, 0, 1 * time.Minute}
}

// ServeHTTP implements the handler function
func (e *Emojify) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	done := e.logger.EmojifyHandlerCalled(r)

	// check the post body
	data, err := e.checkPostBody(r)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
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
		rw.Write([]byte(key))
		done(http.StatusOK, nil)
		return
	}

	// add to a queue to save overloading the service
	st := time.Now()
	for e.currentWorkers() >= e.faceboxWorkers {
		if time.Now().Sub(st) > e.workerTimeout {
			http.Error(rw, "Service Busy", http.StatusTooManyRequests)
			done(http.StatusTooManyRequests, nil)
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	// ok to continue aquire a worker
	e.aquireWorker()

	// check if the request has been cancelled
	if r.Context().Err() != nil {
		// finished release the worker
		e.releaseWorker()
		done(http.StatusGone, errors.New("Client disconnected"))
		return
	}

	// fetch the image
	f, img, err := e.fetchImage(u.String())
	if err != nil {
		e.releaseWorker()
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		done(http.StatusInternalServerError, err)
		return
	}

	// find faces in the image
	faces, err := e.findFaces(u.String(), f)
	if err != nil {
		e.releaseWorker()
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		done(http.StatusInternalServerError, err)
		return
	}

	// process the image and replace faces with emoji
	data, err = e.processImage(u.String(), faces, img)
	if err != nil {
		e.releaseWorker()
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		done(http.StatusInternalServerError, err)
		return
	}

	// save the cache
	err = e.saveCache(u.String(), key, data)
	if err != nil {
		e.releaseWorker()
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		done(http.StatusInternalServerError, err)
		return
	}

	e.releaseWorker()

	// return the image key
	rw.Write([]byte(key))
	done(http.StatusOK, nil)
}

func (e *Emojify) checkPostBody(r *http.Request) ([]byte, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e.logger.EmojifyHandlerNoPostBody()
		return nil, err
	}

	return data, nil
}

func (e *Emojify) validateURL(data []byte) (*url.URL, error) {
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

func (e *Emojify) checkCache(key string) (bool, error) {
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

func (e *Emojify) fetchImage(uri string) (io.ReadSeeker, image.Image, error) {
	fiDone := e.logger.EmojifyHandlerFetchImage(uri)
	f, err := e.fetcher.FetchImage(uri)
	if err != nil {
		fiDone(http.StatusInternalServerError, err)
		return nil, nil, err
	}

	fiDone(http.StatusOK, nil)

	// check image is valid
	img, err := e.fetcher.ReaderToImage(f)
	if err != nil {
		e.logger.EmojifyHandlerInvalidImage(uri, err)
		return nil, nil, err
	}

	return f, img, nil
}

func (e *Emojify) findFaces(uri string, r io.ReadSeeker) ([]facebox.Face, error) {
	ffDone := e.logger.EmojifyHandlerFindFaces(uri)
	f, err := e.emojifyer.GetFaces(r)
	if err != nil {
		ffDone(http.StatusInternalServerError, err)
		return nil, err
	}

	ffDone(http.StatusOK, nil)
	return f, nil
}

func (e *Emojify) processImage(uri string, faces []facebox.Face, img image.Image) ([]byte, error) {
	emDone := e.logger.EmojifyHandlerEmojify(uri)
	i, err := e.emojifyer.Emojimise(img, faces)
	if err != nil {
		emDone(http.StatusInternalServerError, err)
		return nil, err
	}
	emDone(http.StatusOK, nil)

	// save the image
	out := new(bytes.Buffer)
	err = jpeg.Encode(out, i, &jpeg.Options{Quality: 60})
	if err != nil {
		e.logger.EmojifyHandlerImageEncodeError(uri, err)
		return nil, err
	}

	return out.Bytes(), nil
}

func (e *Emojify) saveCache(uri, key string, data []byte) error {
	cpDone := e.logger.EmojifyHandlerCachePut(uri)

	_, err := e.cache.Put(
		context.Background(),
		&cache.CacheItem{Id: key, Data: data},
	)

	if err != nil {
		cpDone(http.StatusInternalServerError, err)
		return err
	}

	cpDone(http.StatusOK, nil)
	return nil
}

func (e *Emojify) currentWorkers() int32 {
	return atomic.LoadInt32(&e.activeWorkers)
}

func (e *Emojify) aquireWorker() {
	atomic.AddInt32(&e.activeWorkers, 1)
}

func (e *Emojify) releaseWorker() {
	atomic.AddInt32(&e.activeWorkers, -1)
}

// hashFilename returns a md5 hash of the filename
func hashFilename(f string) string {
	h := md5.New()
	io.WriteString(h, f)

	return fmt.Sprintf("%x", h.Sum(nil))
}
