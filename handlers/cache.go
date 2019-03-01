package handlers

import (
	"context"
	"net/http"

	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Cache returns images from the cache
type Cache struct {
	logger logging.Logger
	cache  cache.CacheClient
}

// NewCache creates a new http.Handler for dealing with cache requests
func NewCache(l logging.Logger, c cache.CacheClient) *Cache {
	return &Cache{l, c}
}

// ServeHTTP handles requests for cache
func (c *Cache) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	done := c.logger.CacheHandlerCalled(r)
	vars := mux.Vars(r) // Get varaibles from the request path

	// check the parameters contains a valid url
	f := vars["file"]
	if f == "" {
		c.logger.CacheHandlerBadRequest()
		done(http.StatusBadRequest, nil)

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch the file from the cache
	cgd := c.logger.CacheHandlerGetFile(f)
	d, err := c.cache.Get(context.Background(), &wrappers.StringValue{Value: f})

	if s := status.Convert(err); s != nil && s.Code() == codes.NotFound {
		cgd(http.StatusNotFound, nil)
		done(http.StatusNotFound, nil)

		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if err != nil {
		cgd(http.StatusInternalServerError, err)
		done(http.StatusInternalServerError, nil)

		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	cgd(http.StatusOK, nil)

	fileType := http.DetectContentType(d.Data)

	// all ok return the file
	rw.Header().Add("content-type", fileType)
	rw.Write(d.Data)
	done(http.StatusOK, nil)
}
