package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
)

func statusHandler(responseWriter http.ResponseWriter, request *http.Request) {
	streamKey := request.URL.Query().Get("key")

	if streamKey == "" {
		sessionStatusesHandler(responseWriter, request)
	} else {
		streamStatusHandler(responseWriter, request)
	}

	responseWriter.Header().Add("Content-Type", "application/json")
}

func streamStatusHandler(responseWriter http.ResponseWriter, request *http.Request) {
	streamKey := request.URL.Query().Get("key")

	session, ok := manager.SessionsManager.GetSessionByID(streamKey)

	if !ok {
		slog.Info("Could not find active stream", "streamKey", streamKey)
		helpers.LogHTTPError(
			responseWriter,
			"No active stream found",
			http.StatusNotFound)

		return
	}

	statusResult := session.GetStreamStatus()

	if err := json.NewEncoder(responseWriter).Encode(statusResult); err != nil {
		helpers.LogHTTPError(
			responseWriter,
			"Internal Server Error",
			http.StatusInternalServerError)
		slog.Error("API.Status Error", "err", err)
	}

	responseWriter.Header().Add("Content-Type", "application/json")
}

func sessionStatusesHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == "DELETE" {
		return
	}

	if isDisabled := os.Getenv(environment.DisableStatus); isDisabled != "" {
		helpers.LogHTTPError(
			responseWriter,
			"Status Service Unavailable",
			http.StatusServiceUnavailable)

		return
	}

	if err := json.NewEncoder(responseWriter).Encode(manager.SessionsManager.GetSessionStates(false)); err != nil {
		helpers.LogHTTPError(
			responseWriter,
			"Internal Server Error",
			http.StatusInternalServerError)

		slog.Error("Internal Server Error", "err", err)
	}

	responseWriter.Header().Add("Content-Type", "application/json")
}
