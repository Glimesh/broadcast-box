package whip

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
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

		if err != nil {
			helpers.LogHTTPError(
				responseWriter,
				"Profile not found",
				http.StatusNoContent)

			return
		}

		if err := json.NewEncoder(responseWriter).Encode(profile); err != nil {
			helpers.LogHTTPError(
				responseWriter,
				"Internal Server Error",
				http.StatusInternalServerError)
			slog.Error("Profile Encode Error", "err", err)
		}

		responseWriter.Header().Add("Content-Type", "application/json")
	}

	// Update Profile
	if request.Method == "POST" {
		slog.Info("Updating Profile")

		body, _ := io.ReadAll(request.Body)
		var payload updateProfilePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			helpers.LogHTTPError(
				responseWriter,
				"Internal Server Error",
				http.StatusInternalServerError)
			slog.Error("Profile Update Error", "err", err)
			return
		}

		// Update stored profile
		err := authorization.UpdateProfile(token, payload.Motd, payload.IsPublic)
		if err != nil {
			helpers.LogHTTPError(
				responseWriter,
				"Internal Server Error",
				http.StatusInternalServerError)
			slog.Error("Profile Update Error", "err", err)
			return
		}

		profile, _ := authorization.GetPersonalProfile(token)

		// Update current session
		manager.SessionsManager.UpdateProfile(profile)
	}

	responseWriter.Header().Add("Content-Type", "application/json")
}
