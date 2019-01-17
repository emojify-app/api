package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/handlers"
	"github.com/emojify-app/api/logging"
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
var statsDServer = flag.String("statsd-server", "localhost:8125", "StatsD server location")
var audience = flag.String("authn-audience", "emojify", "AuthN audience")
var disableAuth = flag.Bool("authn-disable", false, "Disable authn integration")
var bindAddress = flag.String("bind-address", "localhost:9090", "Bind address for the server defaults to localhost:9090")
var path = flag.String("path", "/", "Path to mount API, defaults to /")
var cacheAddress = flag.String("cache-address", "localhost", "Address for the Cache service")
var paymentGatewayURI = flag.String("payment-address", "localhost", "Address for the Payment gateway service")
var logFormat = flag.String("log_format", "text", "Log output format [text,json]")

func main() {
	flag.Parse()

	logger, err := logging.New("api", *statsDServer, "DEBUG", *logFormat)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger.ServiceStart("localhost", "9090")
	logger.Log().Info(
		"Startup parameters",
		"cache_type", *cacheType,
		"statsDServer", *statsDServer,
		"allowedOrigin", *allowedOrigin,
	)

	r := mux.NewRouter()
	// update the path
	if !strings.HasSuffix(*path, "/") {
		*path = *path + "/"
	}

	authRouter := r.PathPrefix(*path).Subrouter()
	router := r.PathPrefix(*path).Subrouter()

	var cache emojify.Cache
	if *cacheType == "redis" {
		cache = emojify.NewRedisCache(*redisLocation)
	} else {
		cache = emojify.NewFileCache("./cache/")
	}

	f := &emojify.FetcherImpl{}
	e := emojify.NewEmojify(f, "./images/")

	ch := handlers.NewCache(logger, cache)
	router.Handle("/cache", ch).Methods("GET")

	hh := handlers.NewHealth(logger)
	router.Handle("/health", hh).Methods("GET")

	eh := handlers.NewEmojify(e, f, logger, cache)
	authRouter.Handle("/", eh).Methods("POST")
	//ph := paymentHandler{logger: logger.Named("paymentHandler"), paymentGatewayURI: *paymentGatewayURI}
	//mux.HandleFunc(*path+"payments", ph.ServeHTTP)

	// If auth is disabled do not use JWT auth
	if !*disableAuth {
		m, err := handlers.NewJWTAuthMiddleware(*authNServer, *audience, logger.Log(), eh)
		if err != nil {
			logger.Log().Error("Unable to create JWT Auth Middleware", "error", err)
			os.Exit(1)
		}

		authRouter.Use(m.Middleware)
	}

	// setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{*allowedOrigin},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})
	handler := c.Handler(r)

	err = http.ListenAndServe(*bindAddress, handler)
	logger.Log().Error("Unable to start server", "error", err)
}
