package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/nicholasjackson/env"
	"google.golang.org/grpc"

	"github.com/emojify-app/api/handlers"
	"github.com/emojify-app/api/logging"
	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/protos/emojify"
	"github.com/rs/cors"
	//_ "net/http/pprof"
)

func init() {
	http.DefaultClient.Timeout = 3000 * time.Millisecond
}

var version = "dev"

var bindAddress = env.String("BIND_ADDRESS", false, "localhost:9090", "Bind address for the server defaults to localhost:9090")
var path = env.String("API_PATH", false, "/", "Path to mount API, defaults to /")

// authentication flags
var allowedOrigin = env.String("ALLOW_ORIGIN", false, "*", "CORS origin")

// external service flags
var statsDServer = env.String("STATSD_SERVER", false, "localhost:8125", "StatsD server location")
var emojifyAddress = env.String("EMOJIFY_ADDRESS", false, "localhost", "Address for the Emojify service")
var cacheAddress = env.String("CACHE_ADDRESS", false, "localhost", "Address for the Cache service")
var paymentGatewayURI = env.String("PAYMENT_ADDRESS", false, "localhost", "Address for the Payment gateway service")

// logging settings
var logFormat = env.String("LOG_FORMAT", false, "text", "Log output format [text,json]")
var logLevel = env.String("LOG_LEVEL", false, "info", "Log output level [trace,info,debug,warn,error]")

// performance testing flags
// these flags allow the user to inject faults into the service for testing purposes
var cacheErrorRate = env.Float64("CACHE_ERROR_RATE", false, 0.0, "Percentage where cache handler will report an error")
var cacheErrorType = env.String("CACHE_ERROR_TYPE", false, "http_error", "Type of error [http_error, delay]")
var cacheErrorCode = env.Int("CACHE_ERROR_CODE", false, http.StatusInternalServerError, "Error code to return on error")
var cacheErrorDelay = env.Duration("CACHE_ERROR_DELAY", false, 0*time.Second, "Error delay [1s,100ms]")

var help = flag.Bool("help", false, "--help to show help")

func main() {
	flag.Parse()
	env.Parse()

	// if the help flag is passed show configuration options
	if *help == true {
		fmt.Println("Emojify API version:", version)
		fmt.Println("Configuration values are set using environment variables, for info please see the following list:")
		fmt.Println("")
		fmt.Println(env.Help())
		os.Exit(0)
	}

	// configure the logger
	logger, err := logging.New("api", version, *statsDServer, *logLevel, *logFormat)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger.ServiceStart(*bindAddress, version)
	logger.Log().Info(
		"Startup parameters",
		"statsDServer", *statsDServer,
		"allowedOrigin", *allowedOrigin,
	)

	// if the user has configured a path, make sure it ends in a /
	if !strings.HasSuffix(*path, "/") {
		*path = *path + "/"
	}

	logger.Log().Info("Api listening on", "path", *path)

	// create the cache client
	logger.Log().Info("Connecting to cache", "address", *cacheAddress)
	cacheConn, err := grpc.Dial(*cacheAddress, grpc.WithInsecure())
	if err != nil {
		logger.Log().Error("Unable to create cache gRPC client", err)
		os.Exit(1)
	}
	cacheClient := cache.NewCacheClient(cacheConn)

	// create the emojify client
	logger.Log().Info("Connecting to emojify", "address", *emojifyAddress)
	emojifyConn, err := grpc.Dial(*emojifyAddress, grpc.WithInsecure())
	if err != nil {
		logger.Log().Error("Unable to create emojify gRPC client", err)
		os.Exit(1)
	}
	emojifyClient := emojify.NewEmojifyClient(emojifyConn)

	// create handlers
	hh := handlers.NewHealth(logger, emojifyClient, cacheClient)
	ch := handlers.NewCache(logger, cacheClient)
	ehp := handlers.NewEmojifyPost(logger, emojifyClient)
	ehg := handlers.NewEmojifyGet(logger, emojifyClient)

	// configure routing
	r := mux.NewRouter()

	// add profiling
	// r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	baseRouter := r.PathPrefix(*path).Subrouter()                // base subrouter with no middleware
	cacheRouter := r.PathPrefix(*path + "cache").Subrouter()     // caching subrouter
	emojifyRouter := r.PathPrefix(*path + "emojify").Subrouter() // caching subrouter

	baseRouter.Handle("/health", hh).Methods("GET")
	emojifyRouter.Handle("/", ehp).Methods("POST")
	emojifyRouter.Handle("/{id}", ehg).Methods("GET")
	cacheRouter.Handle("/{id}", ch).Methods("GET")

	// Setup error injection for testing
	if *cacheErrorRate != 0.0 {
		logger.Log().Info("Injecting errors into cache handler", "rate", *cacheErrorRate, "code", *cacheErrorCode)

		em := handlers.NewErrorMiddleware(*cacheErrorRate, *cacheErrorCode, *cacheErrorDelay, *cacheErrorType, logger)
		cacheRouter.Use(em.Middleware)

		logger.Log().Info("Injecting errors into emojify handler", "rate", *cacheErrorRate, "code", *cacheErrorCode)
		em2 := handlers.NewErrorMiddleware(*cacheErrorRate, *cacheErrorCode, *cacheErrorDelay, *cacheErrorType, logger)
		emojifyRouter.Use(em2.Middleware)
	}

	// setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{*allowedOrigin},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		Debug:            false,
	})

	handler := c.Handler(r)

	logger.Log().Info("Starting server")
	err = http.ListenAndServe(*bindAddress, handler)
	logger.Log().Error("Unable to start server", "error", err)
}
