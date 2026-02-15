package server

import (
	"log"
	"net/http"
	"os"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/handlers"
)

var (
	defaultHttpAddress         string = ":8080"
	defaultHttpRedirectAddress string = ":80"
)

func startHttpServer(serverMux http.HandlerFunc) {
	server := &http.Server{
		Handler: serverMux,
		Addr:    getHttpAddress(),
	}

	log.Println("Starting HTTP server at", getHttpAddress())
	log.Fatal(server.ListenAndServe())
}

func getHttpAddress() string {
	if httpAddress := os.Getenv(environment.HTTP_ADDRESS); httpAddress != "" {
		return httpAddress
	}

	return defaultHttpAddress
}

func setupHttpRedirect() {
	if shouldRedirectToHttps := os.Getenv(environment.HTTP_ENABLE_REDIRECT); shouldRedirectToHttps != "" {
		httpRedirectPort := defaultHttpRedirectAddress

		if httpRedirectPortEnvVar := os.Getenv(environment.HTTPS_REDIRECT_PORT); httpRedirectPortEnvVar != "" {
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
