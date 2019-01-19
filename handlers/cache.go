package handlers

import (
	"net/http"

	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/logging"
	"github.com/gorilla/mux"
)

// Cache returns images from the cache
type Cache struct {
	logger logging.Logger
	cache  emojify.Cache
}

// NewCache creates a new http.Handler for dealing with cache requests
func NewCache(l logging.Logger, c emojify.Cache) *Cache {
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
	d, err := c.cache.Get(f)
	if err != nil {
		cgd(http.StatusInternalServerError, err)
		done(http.StatusNotFound, nil)

		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if d == nil || len(d) == 0 {
		cgd(http.StatusNotFound, nil)
		done(http.StatusNotFound, nil)

		rw.WriteHeader(http.StatusNotFound)
		return
	}

	cgd(http.StatusOK, nil)

	fileType := http.DetectContentType(d)

	// all ok return the file
	rw.Header().Add("content-type", fileType)
	rw.Write(d)
	done(http.StatusOK, nil)
}
