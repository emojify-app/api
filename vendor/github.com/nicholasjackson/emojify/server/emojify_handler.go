package main

import (
	"fmt"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/nicholasjackson/emojify/emojify"
)

type emojiHandler struct {
	emojifyer emojify.Emojify
	fetcher   emojify.Fetcher
}

func (e *emojiHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var u *url.URL
	var err error

	if u, err = validateURL(r); err != nil {
		log.Println("Unable to process url: %v", err)
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(err.Error()))
		return
	}

	f, err := e.fetcher.FetchImage(u.String())
	if err != nil {
		log.Println(fmt.Errorf("Unable to fetch image %v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "unable to fetch image %v", u.String())
		return
	}

	img, err := e.fetcher.ReaderToImage(f)
	if err != nil {
		log.Println(fmt.Errorf("invalid image format %v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "invalid image format %v", u.String())
		return
	}

	faces, err := e.emojifyer.GetFaces(f)
	if err != nil {
		log.Println(fmt.Errorf("Unable to find faces %v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "unable to find faces in %v", u.String())
		return
	}

	i, err := e.emojifyer.Emojimise(img, faces)
	if err != nil {
		log.Println(fmt.Errorf("Unable to emojify %v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "unable to emojify %v", u.String())
		return
	}

	rw.Header().Add("content-type", "image/png")
	png.Encode(rw, i)
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
