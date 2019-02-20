package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"github.com/emojify-app/api/emojify"
	"github.com/emojify-app/api/handlers"
	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/rs/cors"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func init() {
	http.DefaultClient.Timeout = 3000 * time.Millisecond
}

var version = "dev"

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

// performance testing flags
// these flags allow the user to inject faults into the service for testing purposes
var cacheErrorRate = flag.Float64("cache-error-rate", 0.0, "Percentage where cache handler will report an error")
var cacheErrorCode = flag.Int("cache-error-code", http.StatusInternalServerError, "Error code to return on error")

func main() {
	flag.Parse()

	logger, err := logging.New("api", version, *statsDServer, "DEBUG", *logFormat)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger.ServiceStart("localhost", "9090", version)
	logger.Log().Info(
		"Startup parameters",
		"statsDServer", *statsDServer,
		"allowedOrigin", *allowedOrigin,
	)

	// if the user has configured a path, make sure it ends in a /
	if !strings.HasSuffix(*path, "/") {
		*path = *path + "/"
	}

	r := mux.NewRouter()
	baseRouter := r.PathPrefix(*path).Subrouter()            // base subrouter with no middleware
	authRouter := r.PathPrefix(*path).Subrouter()            // handlers which require authentication
	cacheRouter := r.PathPrefix(*path + "cache").Subrouter() // caching subrouter

	logger.Log().Info("Connecting to cache", "address", *cacheAddress)

	conn, err := grpc.Dial(*cacheAddress, grpc.WithInsecure())
	if err != nil {
		logger.Log().Error("Unable to create gRPC client", err)
		os.Exit(1)
	}
	cacheClient := cache.NewCacheClient(conn)

	f := &emojify.FetcherImpl{}
	e := emojify.NewEmojify(f, "./images/")

	ch := handlers.NewCache(logger, cacheClient)
	cacheRouter.Handle("/{file}", ch).Methods("GET")

	hh := handlers.NewHealth(logger, e, cacheClient)
	baseRouter.Handle("/health", hh).Methods("GET")

	eh := handlers.NewEmojify(e, f, logger, cacheClient)
	authRouter.Handle("/", eh).Methods("POST")

	// If auth is disabled do not use JWT auth
	if !*disableAuth {
		m, err := handlers.NewJWTAuthMiddleware(*authNServer, *audience, logger.Log(), eh)
		if err != nil {
			logger.Log().Error("Unable to create JWT Auth Middleware", "error", err)
			os.Exit(1)
		}

		authRouter.Use(m.Middleware)
	}

	// Setup error injection for testing
	if *cacheErrorRate != 0.0 {
		logger.Log().Info("Injecting errors into cache handler", "rate", *cacheErrorRate, "code", *cacheErrorCode)

		em := handlers.NewErrorMiddleware(*cacheErrorRate, *cacheErrorCode, logger)
		cacheRouter.Use(em.Middleware)
	}

	// setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{*allowedOrigin},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		Debug:            false,
	})
	handler := c.Handler(r)

	err = http.ListenAndServe(*bindAddress, handler)
	logger.Log().Error("Unable to start server", "error", err)
}
