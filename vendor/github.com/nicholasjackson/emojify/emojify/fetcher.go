package emojify

import (
	"bytes"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const MaxFileSize = 4000000 // 4MB

type Fetcher interface {
	FetchImage(uri string) (io.ReadSeeker, error)
	ReaderToImage(r io.ReadSeeker) (image.Image, error)
}

type FetcherImpl struct {
	MaxFileSize int
}

func (f *FetcherImpl) FetchImage(uri string) (io.ReadSeeker, error) {
	log.Println("Fetching: ", uri)

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := io.LimitReader(resp.Body, MaxFileSize)
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf), nil
}

func (f *FetcherImpl) ReaderToImage(r io.ReadSeeker) (image.Image, error) {
	_, err := r.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	in, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	return in, nil
}
