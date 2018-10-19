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
var disableAuth = flag.Bool("authn-disable", false, "Disable authn integration")
var bindAddress = flag.String("bind-address", "localhost:9090", "Bind address for the server defaults to localhost:9090")
var path = flag.String("path", "/", "Path to mount API, defaults to /")
var cacheAddress = flag.String("cache-address", "localhost", "Address for the Cache service")
var paymentGatewayURI = flag.String("payment-address", "localhost", "Address for the Payment gateway service")

func main() {
	flag.Parse()

	logger := hclog.Default()
	logger.Info("Started API Server", "version", "0.5.4")
	logger.Info("Setting allowed origin for CORS", "origin", *allowedOrigin)

	mux := http.NewServeMux()
	// update the path
	if !strings.HasSuffix(*path, "/") {
		*path = *path + "/"
	}

	var cache emojify.Cache
	if *cacheType == "redis" {
		cache = emojify.NewRedisCache(*redisLocation)
	} else {
		cache = emojify.NewFileCache("./cache/")
	}

	f := &emojify.FetcherImpl{}
	e := emojify.NewEmojify(f, "./images/")

	ch := CacheHandler{logger: logger.Named("cacheHandler"), cache: cache}
	mux.HandleFunc(*path+"cache", ch.Handle)

	hh := healthHandler{}
	mux.HandleFunc(*path+"health", hh.Handle)

	eh := emojiHandler{fetcher: f, emojifyer: e, logger: logger.Named("emojiHandler"), cache: cache}
	ph := paymentHandler{logger: logger.Named("paymentHandler"), paymentGatewayURI: *paymentGatewayURI}

	// If auth is disabled do not use JWT auth
	if *disableAuth {
		mux.HandleFunc(*path, eh.Handle)
		mux.HandleFunc(*path+"payments", ph.ServeHTTP)
	} else {
		ah, err := NewJWTAuthMiddleware(*authNServer, *audience, logger, eh)
		if err != nil {
			logger.Error("Unable to create JWT Auth Middleware", "error", err)
			log.Fatal(err)
		}

		ph, err := NewJWTAuthMiddleware(*authNServer, *audience, logger, ph)
		if err != nil {
			logger.Error("Unable to create JWT Auth Middleware", "error", err)
			log.Fatal(err)
		}

		mux.HandleFunc(*path, ah.Handle)
		mux.HandleFunc(*path+"payments", ph.Handle)
	}

	// setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{*allowedOrigin},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})
	handler := c.Handler(mux)

	logger.Info("Starting server on ", *bindAddress)

	err := http.ListenAndServe(*bindAddress, handler)
	log.Fatal(err)
}
