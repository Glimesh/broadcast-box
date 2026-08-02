package admin

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

func LoggingHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if isValidMethod := verifyValidMethod("GET", responseWriter, request); !isValidMethod {
		return
	}

	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHTTPError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	file, err := environment.GetLogFileReader()
	if err != nil {
		slog.Error("API.Admin.Logging Error", "err", err)
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	if _, err := io.Copy(responseWriter, file); err != nil {
		slog.Error("API.Admin.Logging: Error writing file to response", "err", err)
		helpers.LogHTTPError(responseWriter, "Invalid request", http.StatusBadRequest)
	}

	err = file.Close()
	if err != nil {
		slog.Error("API.Admin.Logging Error", "err", err)
	}
}
