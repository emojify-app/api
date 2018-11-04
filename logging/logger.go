package logging

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	hclog "github.com/hashicorp/go-hclog"
)

// Logger defines an interface which handles logging and metrics for the api
type Logger interface {
	CreateCalled(requestID, url string) *Trace
	GetCalled(requestID, url string) *Trace
}

// Trace implements a trace block which is used to notify the logger when a block has completed
type Trace struct {
	Finished func()
}

// LoggerImpl is a concrete implementation of the Logger interface
type LoggerImpl struct {
	serviceName string
	handler     string
	statsd      *statsd.Client
	log         hclog.Logger
}

// NewLogger creates a new logger configured with the info for the service
func NewLogger(statsdAddress, serviceName string, handler string) (Logger, error) {
	log := hclog.New(&hclog.LoggerOptions{
		Name: serviceName,
	})

	sd, err := statsd.New(statsdAddress)
	if err != nil {
		return nil, err
	}

	return &LoggerImpl{
		serviceName: serviceName,
		handler:     handler,
		statsd:      sd,
		log:         log,
	}, nil
}

func (l *LoggerImpl) getName(action string) string {
	return fmt.Sprintf("%s.%s.%s", l.serviceName, l.handler, action)
}

// CreateCalled starts a new trace when the create endpoint is called
func (l *LoggerImpl) CreateCalled(requestID, url string) *Trace {
	n := l.getName("create")

	l.log.Info(n+".called", "rid", requestID, "url", url)

	st := time.Now()
	return &Trace{
		Finished: func() {
			d := time.Now().Sub(st)

			l.statsd.Timing(n, d, []string{"rid:" + requestID}, 1.0)
			l.log.Info(n+".finished", "rid", requestID, "duration", d)
		},
	}
}

// GetCalled starts a new trace when the create endpoint is called
func (l *LoggerImpl) GetCalled(requestID, url string) *Trace {
	n := l.getName("get")

	l.log.Info(n+".called", "rid", requestID, "url", url)

	st := time.Now()
	return &Trace{
		Finished: func() {
			d := time.Now().Sub(st)

			l.statsd.Timing(n, d, []string{"rid:" + requestID}, 1.0)
			l.log.Info(n+".finished", "rid", requestID, "duration", d)
		},
	}
}
