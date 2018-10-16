package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"strings"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/emojify-api/emojify"
	"github.com/rs/cors"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func init() {
	http.DefaultClient.Timeout = 3000 * time.Millisecond
}

var cacheType = flag.String("cache-type", "file", "Cache type, redis/file")
var redisLocation = flag.String("redis-location", "localhost:1234", "Location for the redis server")
var redisPassword = flag.String("redis-password", "", "Password for redis server")
var allowedOrigin = flag.String("allow-origin", "*", "CORS origin")
var authNServer = flag.String("authn-server", "http://localhost:3000", "AuthN server location")
var audience = flag.String("authn-audience", "emojify", "AuthN audience")
var bindAddress = flag.String("bind-address", "localhost:9090", "Bind address for the server defaults to localhost:9090")
var path = flag.String("path", "/", "Path to mount API, defaults to /")

func main() {
	flag.Parse()

	logger := hclog.Default()
	logger.Info("Started API Server", "version", "0.5.3")
	logger.Info("Setting allowed origin for CORS", "origin", *allowedOrigin)

	var cache emojify.Cache
	if *cacheType == "redis" {
		cache = emojify.NewRedisCache(*redisLocation)
	} else {
		cache = emojify.NewFileCache("./cache/")
	}

	f := &emojify.FetcherImpl{}
	e := emojify.NewEmojify(f, "./images/")
	eh := emojiHandler{fetcher: f, emojifyer: e, logger: logger.Named("emojiHandler"), cache: cache}
	ch := CacheHandler{logger: logger.Named("cacheHandler"), cache: cache}
	hh := healthHandler{}
	ah, err := NewJWTAuthMiddleware(*authNServer, *audience, logger, eh)

	if err != nil {
		logger.Error("Unable to create JWT Auth Middleware", "error", err)
		log.Fatal(err)
	}

	// update the path
	if !strings.HasSuffix(*path, "/") {
		*path = *path + "/"
	}

	mux := http.NewServeMux()
	mux.HandleFunc(*path, ah.Handle)
	mux.HandleFunc(*path+"health", hh.Handle)
	mux.HandleFunc(*path+"cache", ch.Handle)

	// setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{*allowedOrigin},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})
	handler := c.Handler(mux)

	logger.Info("Starting server on ", *bindAddress)

	err = http.ListenAndServe(*bindAddress, handler)
	log.Fatal(err)
}
