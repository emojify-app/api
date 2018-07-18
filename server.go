package main

import (
	"log"
	"net/http"
	"time"

	"github.com/nicholasjackson/emojify/emojify"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func init() {
	http.DefaultClient.Timeout = 3000 * time.Millisecond
}

func main() {
	f := &emojify.FetcherImpl{}
	e := emojify.NewEmojify(f, "./images/")
	eh := emojiHandler{fetcher: f, emojifyer: e}
	hh := healthHandler{}

	http.HandleFunc("/", eh.Handle)
	http.HandleFunc("/health", hh.Handle)
	http.Handle("/cache/", http.StripPrefix("/cache", http.FileServer(http.Dir("./cache"))))

	log.Println("Starting server on port 9090")
	err := http.ListenAndServe(":9090", http.DefaultServeMux)

	log.Fatal(err)
}
