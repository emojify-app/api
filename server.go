package main

import (
	"flag"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/env"
)

func init() {
	http.DefaultClient.Timeout = 3000 * time.Millisecond
}

var version = "0.6.0"

var help = flag.Bool("help", false, "--help to show help")

var bindAddress = env.String("BIND_ADDRESS", false, "localhost:9090", "Bind address for the server defaults to localhost:9090")
var allowedOrigin = env.String("ALLOW_ORIGIN", false, "*", "CORS origin")
var cacheAddress = env.String("CACHE_URI", false, "localhost", "Address for the Cache service")
var paymentGatewayURI = env.String("PAYMENT_URI", false, "localhost", "Address for the Payment gateway service")
var authNServer = env.String("AUTHN_URI", false, "http://localhost:3000", "AuthN server location")
var audience = env.String("AUTHN_URI_AUDIENCE", false, "emojify", "AuthN audience")
var disableAuth = env.Bool("AUTHN_DISABLE", false, false, "Disable authn integration")

func main() {
	flag.Parse()

	// if the help flag is passed show configuration options
	if *help == true {
		fmt.Println("API service version:", version)
		fmt.Println("Configuration values are set using environment variables, for info please see the following list")
		fmt.Println("")
		fmt.Println(env.Help())
	}

	// Parse the environment variables and exit on error
	err := env.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger := hclog.Default()
	logger.Info("Started API Server", "version", version)
	logger.Info("Setting allowed origin for CORS", "origin", *allowedOrigin)

	//setupRouter()
}

/*
func setupRouter() mux.Router {

	mux := http.NewServeMux()

	hh := healthHandler{}
	mux.HandleFunc("/health", hh.Handle)

	//eh := emojiHandler{fetcher: f, emojifyer: e, logger: logger.Named("emojiHandler"), cache: cache}
	ph := paymentHandler{logger: logger.Named("paymentHandler"), paymentGatewayURI: *paymentGatewayURI}

	// If auth is disabled do not use JWT auth
	if *disableAuth {
		mux.HandleFunc("/payments", ph.ServeHTTP)
	} else {
		ph, err := NewJWTAuthMiddleware(*authNServer, *audience, logger, ph)
		if err != nil {
			logger.Error("Unable to create JWT Auth Middleware", "error", err)
			log.Fatal(err)
		}

		mux.HandleFunc("/payments", ph.Handle)
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

func createDependencies() (hclog.Logger, emojify.Emojify) {
	logger := hclog.Default()

	return logger, e
}
*/
