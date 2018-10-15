package main

import (
	"bytes"
	"fmt"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/asaskevich/govalidator"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/emojify-api/emojify"
)

type emojiHandler struct {
	emojifyer emojify.Emojify
	fetcher   emojify.Fetcher
	logger    hclog.Logger
	cache     emojify.Cache
}

func (e emojiHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	e.Handle(rw, r)
}

func (e *emojiHandler) Handle(rw http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		e.logger.Info("Method not allowed", "method", r.Method)
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e.logger.Info("No body for POST")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	var u *url.URL

	if u, err = validateURL(data); err != nil {
		e.logger.Error("Unable to validate URI", "error", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	key := emojify.HashFilename(u.String())

	e.logger.Info("Checking cache", "key", key)
	ok, err := e.cache.Exists(key)
	if err != nil {
		e.logger.Error("Unable to check cache", "error", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if ok {
		e.logger.Info("Successfully returned image from cache", "key", key)
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(key))
		return
	}

	e.logger.Info("Fetching image", "URI", u.String())
	f, err := e.fetcher.FetchImage(u.String())
	if err != nil {
		e.logger.Error("Unable to fetch image", "error", err, "URI", u.String())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	img, err := e.fetcher.ReaderToImage(f)
	if err != nil {
		e.logger.Error("invalid image format", "error", err, "URI", u.String())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	faces, err := e.emojifyer.GetFaces(f)
	if err != nil {
		e.logger.Error("Unable to find faces", "error", err, "URI", u.String())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	i, err := e.emojifyer.Emojimise(img, faces)
	if err != nil {
		e.logger.Error("Unable to emojify", "error", err, "URI", u.String())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	e.logger.Info("Successfully processed image", "URI", u.String())

	// save the image
	out := new(bytes.Buffer)

	err = png.Encode(out, i)
	if err != nil {
		e.logger.Error("Unable to encode file as png", "URI", u.String(), "error", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// save the cache
	err = e.cache.Put(key, out.Bytes())
	if err != nil {
		e.logger.Error("Unable to cache image", "URI", u.String(), "key", key, "error", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	e.logger.Info("Written file to cache", "URI", u.String(), "key", key, "error", err)

	rw.Write([]byte(key))
}

func validateURL(data []byte) (*url.URL, error) {
	valid := govalidator.IsRequestURL(string(data))
	if valid == false {
		return nil, fmt.Errorf("%v is not a valid URL", string(data))
	}

	u, err := url.ParseRequestURI(string(data))
	if err != nil {
		return nil, fmt.Errorf("unable to parse %v", string(data))
	}

	return u, nil
}
