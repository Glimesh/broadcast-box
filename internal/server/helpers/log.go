package helpers

import (
	"log"
	"net/http"
	"strings"
)

func LogHttpError(responseWriter http.ResponseWriter, error string, code int) {
	log.Println(error)
	http.Error(responseWriter, error, code)
}

func ResolveBearerToken(authHeader string) string {
	const bearerPrefix = "Bearer"
	if result, ok := strings.CutPrefix(authHeader, bearerPrefix); ok {
		return strings.ReplaceAll(result, " ", "")
	}

	return ""
}
