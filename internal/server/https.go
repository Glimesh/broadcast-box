package server

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"os"

	"github.com/glimesh/broadcast-box/internal/environment"
)

var (
	defaultHTTPSAddress string = ":443"
)

func startHTTPSServer(serverMux http.HandlerFunc) {
	sslKey := os.Getenv(environment.SSLKey)
	sslCert := os.Getenv(environment.SSLCert)

	if sslKey == "" {
		slog.Error("Missing SSL Key")
		os.Exit(1)
	}
	if sslCert == "" {
		slog.Error("Missing SSL Certificate")
		os.Exit(1)
	}

	server := &http.Server{
		Handler: serverMux,
		Addr:    getHTTPSAddress(),
	}

	cert, err := tls.LoadX509KeyPair(sslCert, sslKey)
	if err != nil {
		slog.Error("Failed to load X509 key pair", "err", err)
		os.Exit(1)
	}

	server.TLSConfig = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}

	slog.Info("Serving HTTPS server", "address", getHTTPSAddress())
	if err := server.ListenAndServeTLS("", ""); err != nil {
		slog.Error("HTTPS Server error", "err", err)
		os.Exit(1)
	}
}

func getHTTPSAddress() string {

	if httpsAddress := os.Getenv(environment.HTTPAddress); httpsAddress != "" {
		return httpsAddress
	}

	return defaultHTTPSAddress
}
