package server

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

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

	slog.Info("Starting HTTP", "address", getHTTPAddress())
	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server closed with error", "err", err)
		os.Exit(1)
	}
}

func getHTTPAddress() string {
	if httpAddress := os.Getenv(environment.HTTPAddress); httpAddress != "" {
		return httpAddress
	}

	return defaultHTTPAddress
}

func getHTTPRedirectAddress() string {
	address := os.Getenv(environment.HTTPSRedirectPort)
	if address == "" {
		return defaultHTTPRedirectAddress
	}

	if !strings.HasPrefix(address, ":") {
		address = ":" + address
	}

	return address
}

func setupHTTPRedirect() {
	if shouldRedirectToHTTPS := os.Getenv(environment.HTTPEnableRedirect); shouldRedirectToHTTPS != "" {
		go func() {
			slog.Info("Setting up HTTP Redirecting")

			redirectServer := &http.Server{
				Addr:    getHTTPRedirectAddress(),
				Handler: http.HandlerFunc(handlers.RedirectToHttpsHandler),
			}

			slog.Info("Forwarding requests to HTTPS server", "address", redirectServer.Addr)
			err := redirectServer.ListenAndServe()

			if err != nil {
				slog.Error("Redirect Server closed with error", "err", err)
				os.Exit(1)
			}
		}()
	}
}
