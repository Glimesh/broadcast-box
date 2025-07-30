package server

import (
	"log"
	"net/http"
	"os"

	"github.com/glimesh/broadcast-box/internal/server/handlers"
)

func startHttpServer(serverMux http.HandlerFunc) {
	server := &http.Server{
		Handler: serverMux,
		Addr:    getHttpAddress(),
	}

	log.Println("Starting HTTP server at", getHttpAddress())
	server.ListenAndServe()
}

func getHttpAddress() string {

	if httpAddress := os.Getenv("HTTP_ADDRESS"); httpAddress != "" {
		return httpAddress + ":" + getHttpPort()
	}

	return ":" + getHttpPort()
}

func getHttpPort() string {
	if httpPort := os.Getenv("HTTP_PORT"); httpPort != "" {
		return httpPort
	}

	return "80"
}

func setupHttpRedirect() {
	if shouldRedirectToHttps := os.Getenv("HTTP_ENABLE_REDIRECT"); shouldRedirectToHttps != "" {
		go func() {
			log.Println("Setting up HTTP Redirecting")

			httpPort := getHttpPort()

			redirectServer := &http.Server{
				Addr:    ":" + httpPort,
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
