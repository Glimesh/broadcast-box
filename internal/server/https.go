package server

import (
	"crypto/tls"
	"log"
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
		log.Fatal("Missing SSL Key")
	}
	if sslCert == "" {
		log.Fatal("Missing SSL Certificate")
	}

	server := &http.Server{
		Handler: serverMux,
		Addr:    getHTTPSAddress(),
	}

	cert, err := tls.LoadX509KeyPair(sslCert, sslKey)
	if err != nil {
		log.Fatal(err)
	}

	server.TLSConfig = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}

	log.Println("Serving HTTPS server at", getHTTPSAddress())
	log.Fatal(server.ListenAndServeTLS("", ""))
}

func getHTTPSAddress() string {

	if httpsAddress := os.Getenv(environment.HTTPAddress); httpsAddress != "" {
		return httpsAddress
	}

	return defaultHTTPSAddress
}
