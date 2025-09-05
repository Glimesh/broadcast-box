package whip

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/server/helpers"

	"github.com/glimesh/broadcast-box/internal/webrtc/session"
)

type updateProfilePayload struct {
	Motd     string `json:"motd"`
	IsPublic bool   `json:"isPublic"`
}

func ProfileHandler(responseWriter http.ResponseWriter, request *http.Request) {
	token := helpers.ResolveBearerToken(request.Header.Get("Authorization"))

	// Get Profile
	if request.Method == "GET" {
		profile, err := authorization.GetPersonalProfile(token)

		session.WhipSessionsLock.Lock()
		defer session.WhipSessionsLock.Unlock()

		for _, stream := range session.WhipSessions {
			if stream.StreamKey == profile.StreamKey {
				profile.IsActive = stream.HasHost.Load()
			}
		}

		if err != nil {
			helpers.LogHttpError(
				responseWriter,
				"Profile not found",
				http.StatusNoContent)

			return
		}

		if err := json.NewEncoder(responseWriter).Encode(profile); err != nil {
			helpers.LogHttpError(
				responseWriter,
				"Internal Server Error",
				http.StatusInternalServerError)
			log.Println(err.Error())
		}

		responseWriter.Header().Add("Content-Type", "application/json")
	}

	// Update Profile
	if request.Method == "POST" {
		log.Println("Updating Profile")

		body, _ := io.ReadAll(request.Body)
		var payload updateProfilePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			helpers.LogHttpError(
				responseWriter,
				"Internal Server Error",
				http.StatusInternalServerError)
			log.Println("Profile Update Error:", err)
			return
		}

		err := authorization.UpdateProfile(token, payload.Motd, payload.IsPublic)
		if err != nil {
			helpers.LogHttpError(
				responseWriter,
				"Internal Server Error",
				http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}
	}

	responseWriter.Header().Add("Content-Type", "application/json")
}
