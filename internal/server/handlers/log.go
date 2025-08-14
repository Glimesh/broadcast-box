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
	if strings.EqualFold(os.Getenv(environment.LOGGING_API_ENABLED), "true") == false {
		return
	}

	logDir := "logs"
	logFilePath := logDir + "/log"

	file, err := os.Open(logFilePath)
	if err != nil {
		log.Println("API.Log Error:", err)
	}
	defer file.Close()

	responseWriter.Header().Set("Content-Type", "text/plain")

	if _, err := io.Copy(responseWriter, file); err != nil {
		helpers.LogHttpError(responseWriter, "Invalid request", http.StatusBadRequest)
	}
}
