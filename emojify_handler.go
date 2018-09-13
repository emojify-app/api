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
	"github.com/google/uuid"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/emojify-api/emojify"
)

type emojiHandler struct {
	emojifyer emojify.Emojify
	fetcher   emojify.Fetcher
	logger    hclog.Logger
	cache     emojify.Cache
}

func (e *emojiHandler) Handle(rw http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var u *url.URL
	var err error

	if u, err = validateURL(r); err != nil {
		e.logger.Error("Unable to process URI", "error", err)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	e.logger.Info("Checking cache", "URI", u.String())
	ok, err := e.cache.Exists(u.String())
	if err != nil {
		e.logger.Error("Unable to check cache", "error", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if ok {
		e.logger.Info("Successfully returned image from cache", "URI", u.String())
		rw.WriteHeader(http.StatusNotModified)
		rw.Write([]byte(u.String()))
		return
	}

	e.logger.Info("Fetching image", "URI", u.String())
	f, err := e.fetcher.FetchImage(u.String())
	if err != nil {
		e.logger.Error("Unable to fetch image", "error", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	img, err := e.fetcher.ReaderToImage(f)
	if err != nil {
		e.logger.Error("invalid image format", "error", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	faces, err := e.emojifyer.GetFaces(f)
	if err != nil {
		e.logger.Error("Unable to find faces", "error", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	i, err := e.emojifyer.Emojimise(img, faces)
	if err != nil {
		e.logger.Error("Unable to emojify", "error", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := uuid.New().String() + ".png"
	e.logger.Info("Successfully processed image", "file", filename)

	// save the image
	data := []byte{}
	out := bytes.NewBuffer(data)
	png.Encode(out, i)

	// save the cache
	err = e.cache.Put(u.String(), data)
	if err != nil {
		e.logger.Error("Unable to cache image", "file", filename, "error", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write([]byte(u.String()))
}

func validateURL(r *http.Request) (*url.URL, error) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read body")
	}

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
