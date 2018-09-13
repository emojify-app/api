package main

import (
	"log"
	"net/http"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/emojify-api/emojify"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func init() {
	http.DefaultClient.Timeout = 3000 * time.Millisecond
}

func main() {
	logger := hclog.Default()
	logger.Info("Started API Server", "version", 0.3)

	cache := emojify.NewFileCache("./cache/")

	f := &emojify.FetcherImpl{}
	e := emojify.NewEmojify(f, "./images/")
	eh := emojiHandler{fetcher: f, emojifyer: e, logger: logger.Named("emojiHandler"), cache: cache}
	hh := healthHandler{}

	http.HandleFunc("/", eh.Handle)
	http.HandleFunc("/health", hh.Handle)
	http.Handle("/cache/", http.StripPrefix("/cache", http.FileServer(http.Dir("./cache"))))

	logger.Info("Starting server on port 9090")
	err := http.ListenAndServe(":9090", http.DefaultServeMux)

	log.Fatal(err)
}
