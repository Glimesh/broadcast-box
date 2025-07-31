package server

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/server/handlers"
)

func startHttpServer(serverMux http.HandlerFunc) {
	server := &http.Server{
		Handler: serverMux,
		Addr:    getHttpAddress(),
	}

	log.Println("Starting HTTP server at", getHttpAddress())
	err := server.ListenAndServe()
	if err != nil {
		log.Println("Error starting HTTP server", err)
	}
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
	if shouldRedirectToHttps := os.Getenv("HTTP_ENABLE_REDIRECT"); strings.EqualFold(shouldRedirectToHttps, "TRUE") {
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
