package server

import (
	"os"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/handlers"
)

// HTTP Setup
func StartWebServer() {
	setupHttpRedirect()

	serverMux := handlers.GetServeMuxHandler()

	if os.Getenv(environment.SSL_KEY) != "" && os.Getenv(environment.SSL_CERT) != "" {
		startHttpsServer(serverMux)
	} else {
		startHttpServer(serverMux)
	}
}
