package handlers

import (
	"net/http"
	"time"

	"github.com/emojify-app/api/logging"
)

// ErrorMiddleware allows errors to be injected into handlers
type ErrorMiddleware struct {
	logger          logging.Logger
	errorPercentage int
	errorCode       int
	errorType       string
	errorDelay      time.Duration
	requestCount    int
}

// NewErrorMiddleware creates a new ErrorMiddleWare
// errorType = [delay,http_error]
func NewErrorMiddleware(errorPercentage float64, errorCode int, errorDelay time.Duration, errorType string, l logging.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{
		logger:          l,
		errorPercentage: int(1 / errorPercentage),
		errorCode:       errorCode,
		errorDelay:      errorDelay,
		errorType:       errorType,
	}
}

// Middleware is used by gorilla/mux to create a middleware
func (j *ErrorMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		j.requestCount++
		// if request count is greater than the max value for an integer reset
		if j.requestCount == int(^uint(0)>>1) {
			j.requestCount = 1
		}

		// calculate if we need to throw an error or continue as normal
		if j.requestCount%j.errorPercentage == 0 {
			j.logger.ErrorInjectionHandlerError(j.requestCount, j.errorPercentage, j.errorType)

			// is our error a delay or a timeout
			if j.errorType == "http_error" {
				// http error
				http.Error(rw, "Error serving request", j.errorCode)
				return // do not call next
			}

			// delay
			time.Sleep(j.errorDelay)

		}

		next.ServeHTTP(rw, r)
	})
}
