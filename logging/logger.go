package logging

import (
	"fmt"
	"net/http"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
)

var statsPrefix = "api.service."

// Logger defines an interface for common logging operations
type Logger interface {
	ServiceStart(address, port string)

	HealthHandlerCalled() Finished

	CacheHandlerCalled(r *http.Request) Finished
	CacheHandlerBadRequest()
	CacheHandlerFileNotFound(f string)
	CacheHandlerGetFile(f string) Finished

	EmojifyHandlerCalled(r *http.Request) Finished
	EmojifyHandlerNoPostBody()
	EmojifyHandlerInvalidURL(uri string, err error)
	EmojifyHandlerCacheCheck(key string) Finished
	EmojifyHandlerFetchImage(uri string) Finished
	EmojifyHandlerInvalidImage(uri string, err error)
	EmojifyHandlerFindFaces(uri string) Finished
	EmojifyHandlerEmojify(uri string) Finished
	EmojifyHandlerImageEncodeError(uri string, err error)
	EmojifyHandlerCachePut(uri string) Finished

	Log() hclog.Logger
}

// Finished defines a function to be returned by logging methods which contain timers
type Finished func(status int, err error)

// New creates a new logger with the given name and points it at a statsd server
func New(name, statsDServer, logLevel string) (Logger, error) {
	o := hclog.DefaultOptions
	o.Name = name
	o.Level = hclog.LevelFromString(logLevel)
	l := hclog.New(o)

	c, err := statsd.New(statsDServer)

	if err != nil {
		return nil, err
	}

	return &LoggerImpl{l, c}, nil
}

// LoggerImpl is a concrete implementation for the logger function
type LoggerImpl struct {
	l hclog.Logger
	s *statsd.Client
}

// Log returns the underlying logger
func (l *LoggerImpl) Log() hclog.Logger {
	return l.l
}

// ServiceStart logs information about the service start
func (l *LoggerImpl) ServiceStart(address, port string) {
	l.s.Incr(statsPrefix+"started", nil, 1)
	l.l.Info("Service started", "address", address, "port", port)
}

// HealthHandlerCalled logs information when the health handler is called, the returned function
// must be called once work has completed
func (l *LoggerImpl) HealthHandlerCalled() Finished {
	st := time.Now()
	l.l.Info("Health handler called")

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"health.called", time.Now().Sub(st), nil, 1)
		l.l.Info("Health handler finished", "status", status)
	}
}

// CacheHandlerCalled logs information when the cache handler is called, the returned function
// must be called once work has completed
func (l *LoggerImpl) CacheHandlerCalled(r *http.Request) Finished {
	st := time.Now()
	l.l.Info("Cache called", "method", r.Method, "URI", r.URL.String())

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"cache.called", time.Now().Sub(st), getStatusTags(status), 1)
		l.l.Info("Cache handler finished", "status", status)
	}
}

// CacheHandlerBadRequest logs information when the cache handler is called with
// missing parameters
func (l *LoggerImpl) CacheHandlerBadRequest() {
	l.s.Incr(statsPrefix+"cache.missing_parameter", nil, 1)
	l.l.Info("File is a required parameter", "handler", "cache")
}

// CacheHandlerFileNotFound logs information when the a file is missing from the cache
func (l *LoggerImpl) CacheHandlerFileNotFound(f string) {
	l.s.Incr(statsPrefix+"cache.file_not_found", nil, 1)
	l.l.Info("File not found in cache", "handler", "cache", "file", f)
}

// CacheHandlerGetFile logs information when data is fetched from the cache
func (l *LoggerImpl) CacheHandlerGetFile(f string) Finished {
	st := time.Now()
	l.l.Info("Fetching file from cache", "handler", "cache", "file", f)

	return func(status int, err error) {
		if err != nil {
			l.s.Incr(statsPrefix+"cache.error", nil, 1)
			l.l.Error("Error fetching file from cache", "handler", "cache", "file", f, "error", err)
		}

		l.s.Timing(statsPrefix+"cache.get", time.Now().Sub(st), getStatusTags(status), 1)
	}
}

// EmojifyHandlerCalled logs information when the Emojify handler is called
func (l *LoggerImpl) EmojifyHandlerCalled(r *http.Request) Finished {
	st := time.Now()
	l.l.Info("Emojify called", "method", r.Method, "URI", r.URL.String())

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"emojify.called", time.Now().Sub(st), getStatusTags(status), 1)
		l.l.Info("Emojify handler finished", "status", status)
	}
}

// EmojifyHandlerNoPostBody logs information when no post body is sent with the request
func (l *LoggerImpl) EmojifyHandlerNoPostBody() {
	l.l.Info("No body for POST", "handler", "emojify")
	l.s.Incr(statsPrefix+"emojify.no_post_body", nil, 1)
}

// EmojifyHandlerInvalidURL logs information when an invalid URI is passed in the body
func (l *LoggerImpl) EmojifyHandlerInvalidURL(uri string, err error) {
	l.l.Error("Unable to validate URI", "handler", "emojify", "uri", uri, "error", err)
	l.s.Incr(statsPrefix+"emojify.invalid_uri", nil, 1)
}

// EmojifyHandlerCacheCheck logs information about a cache check
func (l *LoggerImpl) EmojifyHandlerCacheCheck(key string) Finished {
	st := time.Now()
	l.l.Info("Checking cache", "handler", "emojify", "key", key)

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"emojify.cache_check", time.Now().Sub(st), getStatusTags(status), 1)
		l.l.Info("Emojify cache finished", "handler", "emojify", "status", status)

		if err != nil {
			l.l.Error("Error checking cache", "handler", "emojify", "key", key, "error", err)
		}
	}
}

// EmojifyHandlerFetchImage logs information about a remote fetch for the image
func (l *LoggerImpl) EmojifyHandlerFetchImage(uri string) Finished {
	st := time.Now()
	l.l.Info("Fetching file", "handler", "emojify", "uri", uri)

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"emojify.fetch_file", time.Now().Sub(st), getStatusTags(status), 1)
		l.l.Info("Fetching file finished", "handler", "emojify", "status", status)

		if err != nil {
			l.l.Error("Error fetching file", "handler", "emojify", "status", status, "uri", uri, "error", err)
		}
	}
}

// EmojifyHandlerInvalidImage logs information when an invalid image is returned from the fetch
func (l *LoggerImpl) EmojifyHandlerInvalidImage(uri string, err error) {
	l.l.Info("Invalid image format", "handler", "emojify", "uri", uri, "error", err)
	l.s.Incr(statsPrefix+"emojify.invalid_image", nil, 1)
}

// EmojifyHandlerFindFaces logs information related to the face lookup call
func (l *LoggerImpl) EmojifyHandlerFindFaces(uri string) Finished {
	st := time.Now()
	l.l.Info("Find faces in image", "handler", "emojify", "uri", uri)

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"emojify.find_faces", time.Now().Sub(st), getStatusTags(status), 1)
		l.l.Info("Find faces finished", "handler", "emojify", "status", status)

		if err != nil {
			l.l.Error("Unable to find faces", "handler", "emojify", "uri", uri, "error", err)
		}
	}
}

// EmojifyHandlerEmojify logs information when emojifying the image
func (l *LoggerImpl) EmojifyHandlerEmojify(uri string) Finished {
	st := time.Now()
	l.l.Info("Emojify image", "handler", "emojify", "uri", uri)

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"emojify.find_faces", time.Now().Sub(st), getStatusTags(status), 1)
		l.l.Info("Find faces finished", "handler", "emojify", "status", status)

		if err != nil {
			l.l.Error("Unable to emojify", "handler", "emojify", "uri", uri, "error", err)
		}
	}
}

// EmojifyHandlerImageEncodeError logs information when an image encode error occurs
func (l *LoggerImpl) EmojifyHandlerImageEncodeError(uri string, err error) {
	l.l.Error("Unable to encode file as png", "handler", "emojify", "uri", uri, "error", err)
	l.s.Incr(statsPrefix+"emojify.image_encode_error", nil, 1)
}

// EmojifyHandlerCachePut logs information when an image is pushed to the cache
func (l *LoggerImpl) EmojifyHandlerCachePut(uri string) Finished {
	st := time.Now()
	l.l.Info("Cache image", "handler", "emojify", "uri", uri)

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"emojify.cache_put", time.Now().Sub(st), getStatusTags(status), 1)
		l.l.Info("Cache image finished", "handler", "emojify", "status", status)

		if err != nil {
			l.l.Error("Unable to save image to cache", "handler", "emojify", "uri", uri, "error", err)
		}
	}
}

func getStatusTags(status int) []string {
	return []string{
		fmt.Sprintf("status:%d", status),
	}
}
