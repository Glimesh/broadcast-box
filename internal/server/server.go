package server

import (
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/server/handlers"
)

// HTTP Setup
func StartWebServer() {
	setupHttpRedirect()

	serverMux := handlers.GetServeMuxHandler()

	isSecureConnection := os.Getenv("USE_SSL")
	if strings.EqualFold(isSecureConnection, "TRUE") {
		startHttpsServer(serverMux)
	} else {
		startHttpServer(serverMux)
	}
}
