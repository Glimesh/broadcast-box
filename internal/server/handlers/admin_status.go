package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
)

func adminStatusHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if isValidMethod := verifyValidMethod("GET", responseWriter, request); !isValidMethod {
		return
	}

	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHttpError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	session.WhipSessionsLock.Lock()
	defer session.WhipSessionsLock.Unlock()

	sessions := session.GetSessionStates(session.WhipSessions, true)
	sessionsCopy := make([]session.StreamSession, len(sessions))
	copy(sessionsCopy, sessions)

	responseWriter.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(responseWriter).Encode(sessionsCopy)
	if err != nil {
		log.Println("API.AdminStatus Error", err)
	}
}
