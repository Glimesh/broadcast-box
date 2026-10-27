package server

import (
	"log"
	"net/http"
	"os"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/handlers"
)

var (
	defaultHTTPAddress         string = ":8080"
	defaultHTTPRedirectAddress string = ":80"
)

func startHTTPServer(serverMux http.HandlerFunc) {
	server := &http.Server{
		Handler: serverMux,
		Addr:    getHTTPAddress(),
	}

	log.Println("Starting HTTP server at", getHTTPAddress())
	log.Fatal(server.ListenAndServe())
}

func getHTTPAddress() string {
	if httpAddress := os.Getenv(environment.HTTPAddress); httpAddress != "" {
		return httpAddress
	}

	return defaultHTTPAddress
}

func setupHTTPRedirect() {
	if shouldRedirectToHTTPS := os.Getenv(environment.HTTPEnableRedirect); shouldRedirectToHTTPS != "" {
		httpRedirectPort := defaultHTTPRedirectAddress

		if httpRedirectPortEnvVar := os.Getenv(environment.HTTPSRedirectPort); httpRedirectPortEnvVar != "" {
			httpRedirectPort = httpRedirectPortEnvVar
		}

		go func() {
			log.Println("Setting up HTTP Redirecting")

			redirectServer := &http.Server{
				Addr:    httpRedirectPort,
				Handler: http.HandlerFunc(handlers.RedirectToHttpsHandler),
			}

			log.Println("Forwarding requests from", redirectServer.Addr, "to HTTPS server")
			err := redirectServer.ListenAndServe()

			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}
