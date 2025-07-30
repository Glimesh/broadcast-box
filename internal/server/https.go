package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
)

func startHttpsServer(serverMux http.HandlerFunc) {
	sslKey := os.Getenv("SSL_KEY")
	sslCert := os.Getenv("SSL_CERT")

	if sslKey == "" {
		log.Fatal("Missing SSL Key")
	}
	if sslCert == "" {
		log.Fatal("Missing SSL Certificate")
	}

	server := &http.Server{
		Handler: serverMux,
		Addr:    getHttpsAddress(),
	}

	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{},
	}

	cert, err := tls.LoadX509KeyPair(sslCert, sslKey)
	if err != nil {
		log.Fatal(err)
	}

	server.TLSConfig.Certificates = append(server.TLSConfig.Certificates, cert)
	log.Println("Serving HTTPS server at", getHttpsAddress())
	log.Fatal(server.ListenAndServeTLS(sslCert, sslKey))
}

func getHttpsPort() string {
	if httpsPort := os.Getenv("HTTPS_PORT"); httpsPort != "" {
		return httpsPort
	}

	return "443"
}

func getHttpsAddress() string {

	if httpsAddress := os.Getenv("HTTPS_ADDRESS"); httpsAddress != "" {
		return httpsAddress
	}

	return ":" + getHttpsPort()
}
