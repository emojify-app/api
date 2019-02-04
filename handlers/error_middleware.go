package handlers

import (
	"net/http"

	"github.com/emojify-app/api/logging"
)

// ErrorMiddleware allows errors to be injected into handlers
type ErrorMiddleware struct {
	logger          logging.Logger
	errorPercentage int
	errorCode       int
	requestCount    int
}

// NewErrorMiddleware creates a new ErrorMiddleWare
func NewErrorMiddleware(errorPercentage float64, errorCode int, l logging.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{l, int(1 / errorPercentage), errorCode, 0}
}

// Middleware is used by gorilla/mux to create a middleware
func (j *ErrorMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		j.requestCount++
		if j.requestCount == int(^uint(0)>>1) {
			j.requestCount = 1
		}

		// calculate if we need to throw an error or continue as normal
		if j.requestCount%j.errorPercentage == 0 {
			j.logger.ErrorInjectionHandlerError(j.requestCount, j.errorPercentage)

			http.Error(rw, "Error serving request", j.errorCode)
			return
		}

		next.ServeHTTP(rw, r)
	})
}
