package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

func logHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if !strings.EqualFold(os.Getenv(environment.LoggingAPIEnabled), "true") {
		return
	}

	if apiKey := os.Getenv(environment.LoggingAPIKey); apiKey != "" {
		authHeader := request.Header.Get("Authorization")
		token := helpers.ResolveBearerToken(authHeader)

		if token == "" {
			helpers.LogHTTPError(responseWriter, "Authorization was invalid", http.StatusUnauthorized)

			return
		} else if token != apiKey {
			helpers.LogHTTPError(responseWriter, "Authorization was invalid", http.StatusUnauthorized)

			return
		}
	}

	file, err := environment.GetLogFileReader()
	if err != nil {
		log.Println("API.Log Error:", err)
	}

	responseWriter.Header().Set("Content-Type", "text/plain")

	if _, err := io.Copy(responseWriter, file); err != nil {
		log.Println("API.Log: Error writing file to response", err)
		helpers.LogHTTPError(responseWriter, "Invalid request", http.StatusBadRequest)
	}

	err = file.Close()
	if err != nil {
		log.Println("API.Log Error:", err)
	}
}
