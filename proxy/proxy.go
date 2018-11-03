package proxy

import (
	"fmt"
	"net/http"
	"time"

	hclog "github.com/hashicorp/go-hclog"
)

// Proxy defines an interface for a proxied request
type Proxy interface {
	Do(r *http.Request) (*http.Response, error)
}

// HTTPProxy is a concrete implementation of Proxy
type HTTPProxy struct {
	logger          hclog.Logger
	upstreamAddress string
	basePath        string
	retryCount      int
	retryDelay      time.Duration
}

// Do sends the request and returns the response
func (h *HTTPProxy) Do(r *http.Request) (*http.Response, error) {
	uri := fmt.Sprintf("%s/", h.upstreamAddress)

	proxyReq, err := http.NewRequest(r.Method, uri, r.Body)
	if err != nil {
		h.logger.Error("Unable to create proxy request", "error", err)
		return nil, err
	}

	proxyReq.Header.Set("Host", r.Host)
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
	//proxyReq.URL.RawQuery = query

	for header, values := range r.Header {
		// ignore cloudflare cf-ray headers
		if header != "cf-ray" {
			for _, value := range values {
				h.logger.Debug("Set request header", "header", header, "value", value)
				proxyReq.Header.Add(header, value)
			}
		}
	}

	//h.logger.Info("Attempting to request from upstream", "upstream", us.Service, "uri", path, "query", query, "method", proxyReq.Method, "protocol", proxyReq.Proto)

	// retry the request 3 times with a backoff
	/*
		var resp *http.Response
		retry := retrier.New(retrier.ConstantBackoff(h.retryCount, h.retryDelay, nil))
		err = retry.Run(func() error {

			var localError error
			resp, localError = r.httpClient.Do(proxyReq)
			if localError != nil {
				h.logger.Error("Unable to contact upstream", "error", localError)
				return localError
			}

			return nil

		})

		if err != nil {
			return nil, err
		}
	*/

	return nil, nil
}
