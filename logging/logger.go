package logging

import (
	"fmt"
	"net/http"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
)

var statsPrefix = "service.api."

// Logger defines an interface for common logging operations
type Logger interface {
	ServiceStart(address, version string)

	HealthHandlerCalled() Finished

	ErrorInjectionHandlerError(requestCount, errorPercentage int, errorType string)

	CacheHandlerCalled(r *http.Request) Finished
	CacheHandlerBadRequest()
	CacheHandlerFileNotFound(f string)
	CacheHandlerGetFile(f string) Finished

	EmojifyHandlerPOSTCalled(r *http.Request) Finished
	EmojifyHandlerGETCalled(r *http.Request) Finished
	EmojifyHandlerNoPostBody()
	EmojifyHandlerInvalidURL(uri string, err error)
	EmojifyHandlerCallCreate(uri string) Finished
	EmojifyHandlerCallQuery(id string) Finished

	Log() hclog.Logger
}

// Finished defines a function to be returned by logging methods which contain timers
type Finished func(status int, err error)

// New creates a new logger with the given name and points it at a statsd server
func New(name, version, statsDServer, logLevel string, logFormat string) (Logger, error) {
	o := &hclog.LoggerOptions{}
	o.Name = name

	// set the log format
	if logFormat == "json" {
		o.JSONFormat = true
	}

	o.Level = hclog.LevelFromString(logLevel)
	l := hclog.New(o)

	c, err := statsd.New(statsDServer)
	c.Tags = []string{fmt.Sprintf("version:%s", version)}

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
func (l *LoggerImpl) ServiceStart(address, version string) {
	l.s.Incr(statsPrefix+"started", nil, 1)
	l.l.Info("Service started", "address", address, "version", version)
}

// HealthHandlerCalled logs information when the health handler is called, the returned function
// must be called once work has completed
func (l *LoggerImpl) HealthHandlerCalled() Finished {
	st := time.Now()
	l.l.Debug("Health handler called")

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"health.called", time.Now().Sub(st), nil, 1)
		if err != nil {
			l.l.Error("Health handler error", "status", status, "error", err)
			return
		}

		l.l.Debug("Health handler finished", "status", status)
	}
}

// ErrorInjectionHandlerError log that an injected error has happened
func (l *LoggerImpl) ErrorInjectionHandlerError(requestCount, errorPercentage int, errorType string) {
	l.l.Error("Injected error", "request count", requestCount, "percentage", errorPercentage, "type", errorType)
	l.s.Incr(statsPrefix+"error.injected", []string{"type:" + errorType}, 1)
}

// CacheHandlerCalled logs information when the cache handler is called, the returned function
// must be called once work has completed
func (l *LoggerImpl) CacheHandlerCalled(r *http.Request) Finished {
	st := time.Now()
	l.l.Debug("Cache called", "method", r.Method, "URI", r.URL.String())

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"cache.called", time.Now().Sub(st), getStatusTags(status), 1)
		if err != nil {
			l.l.Error("Cache handler finished", "status", status, "error", err)
			return
		}

		l.l.Debug("Cache handler finished", "status", status)
	}
}

// CacheHandlerBadRequest logs information when the cache handler is called with
// missing parameters
func (l *LoggerImpl) CacheHandlerBadRequest() {
	l.s.Incr(statsPrefix+"cache.missing_parameter", nil, 1)
	l.l.Error("File is a required parameter", "handler", "cache")
}

// CacheHandlerFileNotFound logs information when the a file is missing from the cache
func (l *LoggerImpl) CacheHandlerFileNotFound(f string) {
	l.s.Incr(statsPrefix+"cache.file_not_found", nil, 1)
	l.l.Debug("File not found in cache", "handler", "cache", "file", f)
}

// CacheHandlerGetFile logs information when data is fetched from the cache
func (l *LoggerImpl) CacheHandlerGetFile(f string) Finished {
	st := time.Now()
	l.l.Debug("Fetching file from cache", "handler", "cache", "file", f)

	return func(status int, err error) {
		if err != nil {
			l.s.Incr(statsPrefix+"cache.error", nil, 1)
			l.l.Error("Error fetching file from cache", "handler", "cache", "file", f, "error", err)
		}

		l.s.Timing(statsPrefix+"cache.get", time.Now().Sub(st), getStatusTags(status), 1)
	}
}

// EmojifyHandlerPOSTCalled logs information when the Emojify handler is called
func (l *LoggerImpl) EmojifyHandlerPOSTCalled(r *http.Request) Finished {
	st := time.Now()
	l.l.Debug("Emojify POST called", "method", r.Method, "URI", r.URL.String())

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"emojify.post.called", time.Now().Sub(st), getStatusTags(status), 1)
		l.l.Debug("Emojify POST handler finished", "status", status)
	}
}

// EmojifyHandlerGETCalled logs information when the Emojify handler is called
func (l *LoggerImpl) EmojifyHandlerGETCalled(r *http.Request) Finished {
	st := time.Now()
	l.l.Debug("Emojify GET called", "method", r.Method, "URI", r.URL.String())

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"emojify.get.called", time.Now().Sub(st), getStatusTags(status), 1)
		l.l.Debug("Emojify GET handler finished", "status", status)
	}
}

// EmojifyHandlerNoPostBody logs information when no post body is sent with the request
func (l *LoggerImpl) EmojifyHandlerNoPostBody() {
	l.l.Error("No body for POST", "handler", "emojify")
	l.s.Incr(statsPrefix+"emojify.no_post_body", nil, 1)
}

// EmojifyHandlerInvalidURL logs information when an invalid URI is passed in the body
func (l *LoggerImpl) EmojifyHandlerInvalidURL(uri string, err error) {
	l.l.Error("Unable to validate URI", "handler", "emojify", "uri", uri, "error", err)
	l.s.Incr(statsPrefix+"emojify.invalid_uri", nil, 1)
}

// EmojifyHandlerCallCreate logs information when the Emojify upstream create method is called
func (l *LoggerImpl) EmojifyHandlerCallCreate(uri string) Finished {
	st := time.Now()
	l.l.Debug("Emojify upstream create called", "URI", uri)

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"emojify.create.called", time.Now().Sub(st), getStatusTags(status), 1)
		l.l.Debug("Emojify upstream create finished", "URI", uri, "status", status)
	}
}

// EmojifyHandlerCallQuery logs information when the Emojify upstream query method is called
func (l *LoggerImpl) EmojifyHandlerCallQuery(id string) Finished {
	st := time.Now()
	l.l.Debug("Emojify upstream query called", "ID", id)

	return func(status int, err error) {
		l.s.Timing(statsPrefix+"emojify.query.called", time.Now().Sub(st), getStatusTags(status), 1)
		l.l.Debug("Emojify upstream query finished", "ID", id, "status", status)
	}
}

func getStatusTags(status int) []string {
	return []string{
		fmt.Sprintf("status:%d", status),
	}
}
