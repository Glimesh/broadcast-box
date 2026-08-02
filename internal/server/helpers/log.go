package helpers

import (
	"log/slog"
	"net/http"
)

func LogHTTPError(responseWriter http.ResponseWriter, error string, code int) {
	slog.Error("HTTP Error", "err", error)
	http.Error(responseWriter, error, code)
}
