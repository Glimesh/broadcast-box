package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
)

func sseHandler(responseWriter http.ResponseWriter, request *http.Request) {
	flusher, ok := responseWriter.(http.Flusher)
	if !ok {
		http.Error(responseWriter, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Add("Content-Type", "text/event-stream")
	responseWriter.Header().Add("Cache-Control", "no-cache")
	responseWriter.Header().Add("Connection", "keep-alive")

	values := strings.Split(request.URL.RequestURI(), "/")
	sessionId := values[len(values)-1]

	debugSseMessages := strings.EqualFold(os.Getenv(environment.DEBUG_PRINT_SSE_MESSAGES), "true")

	ctx := request.Context()

	// Setup WHEP/WHIP session for SSE feed
	sseChannel := getWhipSessionChannel(sessionId)

	if sseChannel == nil {
		sseChannel = getWhepSessionChannel(sessionId)
	}

	if sseChannel == nil {
		helpers.LogHttpError(responseWriter, "Invalid request", http.StatusBadRequest)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sseChannel:
			if debugSseMessages {
				log.Println("API.SSE Sending:", msg)
			}

			if !ok || msg == "close" {
				log.Println("API.SSE: Channel closed")
				return
			}

			if _, err := fmt.Fprintf(responseWriter, "%s\n", msg); err != nil {
				log.Println("API.SSE Error:", err)
			}

			flusher.Flush()
		}
	}
}

func getWhipSessionChannel(sessionId string) chan any {
	var channel chan any
	whipSession, ok := session.SessionManager.GetWhipStreamBySessionId(sessionId)

	if ok {
		channel = whipSession.EventsChannel
	}

	return channel
}

func getWhepSessionChannel(sessionId string) chan any {
	var channel chan any
	whepSession, ok := session.SessionManager.GetWhepStreamBySessionId(sessionId)

	if ok {
		channel = whepSession.SseEventsChannel
	}

	return channel
}
