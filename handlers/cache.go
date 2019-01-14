package handlers

import (
	"net/http"

	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/logging"
)

// CacheHandler returns images from the cache
type Cache struct {
	logger logging.Logger
	cache  emojify.Cache
}

// NewCacheHandler creates a new http.Handler for dealing with cache requests
func NewCache(l logging.Logger, c emojify.Cache) *Cache {
	return &Cache{l, c}
}

// Handle handles requests for cache
func (c *Cache) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	done := c.logger.CacheHandlerCalled(r)

	// check the parameters contains a valid url
	f := r.URL.Query().Get("file")
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

		rw.WriteHeader(http.StatusNotFound)
		return
	}
	cgd(http.StatusNotFound, nil)

	fileType := http.DetectContentType(d)

	// all ok return the file
	rw.Header().Add("content-type", fileType)
	rw.Write(d)
	done(http.StatusOK, nil)
}
