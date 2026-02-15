package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
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
	writeTimeout := 500 * time.Millisecond

	ctx := request.Context()
	responseController := http.NewResponseController(responseWriter)

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
			log.Println("API.SSE: Client disconnected")
			return

		case msg, ok := <-sseChannel:
			if debugSseMessages {
				log.Println("API.SSE Sending:", msg)
			}

			if !ok || msg == "close" {
				log.Println("API.SSE: Channel closed")
				return
			}

			if err := responseController.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil && !errors.Is(err, http.ErrNotSupported) {
				log.Println("API.SSE SetWriteDeadline error:", err)
				return
			}

			_, err := fmt.Fprintf(responseWriter, "%s\n", msg)
			if err == nil {
				flusher.Flush()
			}

			if deadlineErr := responseController.SetWriteDeadline(time.Time{}); deadlineErr != nil && !errors.Is(deadlineErr, http.ErrNotSupported) {
				log.Println("API.SSE ClearWriteDeadline error:", deadlineErr)
				return
			}

			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					log.Println("API.SSE Write timeout")
				} else {
					log.Println("API.SSE Write error:", err)
				}
				return
			}
		}
	}
}

func getWhipSessionChannel(sessionId string) chan any {
	var channel chan any
	whipSession, ok := manager.SessionsManager.GetHostSessionById(sessionId)

	if ok {
		channel = whipSession.EventsChannel
	}

	return channel
}

func getWhepSessionChannel(sessionId string) chan any {
	var channel chan any
	whepSession, ok := manager.SessionsManager.GetWhepSessionById(sessionId)

	if ok {
		channel = whepSession.SseEventsChannel
	}

	return channel
}
