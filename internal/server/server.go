package server

import (
	"os"

	"github.com/glimesh/broadcast-box/internal/server/handlers"
)

// HTTP Setup
func StartWebServer() {
	setupHttpRedirect()

	serverMux := handlers.GetServeMuxHandler()

	isSecureConnection := os.Getenv("USE_SSL")
	if isSecureConnection == "TRUE" {
		startHttpsServer(serverMux)
	} else {
		startHttpServer(serverMux)
	}
}
