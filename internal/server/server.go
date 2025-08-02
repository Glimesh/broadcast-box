package server

import (
	"os"

	"github.com/glimesh/broadcast-box/internal/server/handlers"
)

// HTTP Setup
func StartWebServer() {
	setupHttpRedirect()

	serverMux := handlers.GetServeMuxHandler()

	if os.Getenv("SSL_KEY") != "" && os.Getenv("SSL_CERT") != "" {
		startHttpsServer(serverMux)
	} else {
		startHttpServer(serverMux)
	}
}
