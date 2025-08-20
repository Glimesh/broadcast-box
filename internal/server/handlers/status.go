package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
)

func getStreamKey(request *http.Request) (streamKey string) {
	queries := request.URL.Query()
	streamKey = queries.Get("key")

	return streamKey
}

func statusHandler(responseWriter http.ResponseWriter, request *http.Request) {
	streamKey := getStreamKey(request)

	if streamKey == "" {
		sessionStatusesHandler(responseWriter, request)
	} else {
		streamStatusHandler(responseWriter, request)
	}

	responseWriter.Header().Add("Content-Type", "application/json")
}

func streamStatusHandler(responseWriter http.ResponseWriter, request *http.Request) {
	streamKey := getStreamKey(request)

	session.WhipSessionsLock.Lock()
	defer session.WhipSessionsLock.Unlock()

	stream := session.WhipSessions[streamKey]

	if stream == nil {
		log.Println("Could not find active stream", streamKey)
		helpers.LogHttpError(
			responseWriter,
			"No active stream found",
			http.StatusNotFound)

		return
	}

	statusResult := session.GetStreamStatus(stream)

	if err := json.NewEncoder(responseWriter).Encode(statusResult); err != nil {
		helpers.LogHttpError(
			responseWriter,
			"Internal Server Error",
			http.StatusInternalServerError)
		log.Println(err.Error())
	}

	responseWriter.Header().Add("Content-Type", "application/json")
}

func sessionStatusesHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == "DELETE" {
		return
	}

	if status := os.Getenv(environment.DISABLE_STATUS); status != "" {
		helpers.LogHttpError(
			responseWriter,
			"Status Service Unavailable",
			http.StatusServiceUnavailable)

		return
	}

	session.WhipSessionsLock.Lock()
	defer session.WhipSessionsLock.Unlock()

	if err := json.NewEncoder(responseWriter).Encode(session.GetSessionStates(session.WhipSessions, false)); err != nil {
		helpers.LogHttpError(
			responseWriter,
			"Internal Server Error",
			http.StatusInternalServerError)
		log.Println(err.Error())
	}

	responseWriter.Header().Add("Content-Type", "application/json")
}
