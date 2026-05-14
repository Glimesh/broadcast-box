package helpers

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
)

var debugSessionManager = strings.EqualFold(
	os.Getenv(environment.DebugSessionManager),
	"true",
)

func LogHTTPError(responseWriter http.ResponseWriter, error string, code int) {
	log.Println("LogHTTPError", error)
	http.Error(responseWriter, error, code)
}

// Print Session Manager debug logs
func DebugSessionLog(args ...any) {
	if !debugSessionManager {
		return
	}

	log.Println(append([]any{"[DEBUG]"}, args...)...)
}
