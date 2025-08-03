package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc/status"
)

func statusHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == "DELETE" {
		return
	}

	if status := os.Getenv("DISABLE_STATUS"); status != "" {
		helpers.LogHttpError(
			responseWriter,
			"Status Service Unavailable",
			http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(responseWriter).Encode(status.GetStreamStates()); err != nil {
		helpers.LogHttpError(
			responseWriter,
			"Internal Server Error",
			http.StatusInternalServerError)
		log.Println(err.Error())
	}

	responseWriter.Header().Add("Content-Type", "application/json")
}
