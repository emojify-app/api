package main

import (
	"encoding/base64"
	"net/http"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/emojify-api/emojify"
)

// CacheHandler returns images from the cache
type CacheHandler struct {
	logger hclog.Logger
	cache  emojify.Cache
}

// NewCacheHandler creates a new http.Handler for dealing with cache requests
func NewCacheHandler(l hclog.Logger, c emojify.Cache) *CacheHandler {
	return &CacheHandler{l, c}
}

// Handle handles requests for cache
func (c *CacheHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		c.logger.Info("Method not allowed", "method", r.Method)
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}

	// check the parameters contains a valid url
	f := r.URL.Query().Get("file")
	b, err := base64.StdEncoding.DecodeString(f)
	if err != nil {
		c.logger.Info("File parameter should be base64 encoded", "file", f, "error", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := validateURL(b)
	if err != nil {
		c.logger.Info("File parameter not valid URL", "file", f, "error", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch the file from the cache
	d, err := c.cache.Get(u.String())
	if err != nil {
		c.logger.Info("File not found in cache", "file", f, "error", err)
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	fileType := http.DetectContentType(d)

	c.logger.Info("Found file, returning", "file", u.String())

	// all ok return the file
	rw.Header().Add("content-type", fileType)
	rw.Write(d)
}
