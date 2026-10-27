package whip

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	accountStore "github.com/glimesh/broadcast-box/internal/accounts"
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
		// Try token-based lookup first (for profile-based streams)
		profile, err := authorization.GetPersonalProfile(token)
		if err != nil {
			// Fallback: try stream-key-based lookup (for ad-hoc streams)
			profile, err = authorization.GetPersonalProfileByStreamKey(token)
			if err != nil {
				helpers.LogHTTPError(
					responseWriter,
					"Profile not found",
					http.StatusNoContent)
				return
			}
		}

		// Only reserved profiles (accounts or admin-created profiles) may use
		// profile settings. Anonymous/ad-hoc streams are no longer supported.
		if !authorization.IsProfileReserved(profile.StreamKey) {
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
			log.Println(err.Error())
		}

		responseWriter.Header().Add("Content-Type", "application/json")
	}

	// Update Profile
	if request.Method == "POST" {
		log.Println("Updating Profile")

		// Account stream keys are public channel identifiers. Profile changes
		// must use the account's private stream token instead.
		if accountStore.Exists(token) {
			helpers.LogHTTPError(
				responseWriter,
				"This stream key is reserved; use the private stream token",
				http.StatusUnauthorized)
			return
		}

		body, _ := io.ReadAll(request.Body)
		var payload updateProfilePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			helpers.LogHTTPError(
				responseWriter,
				"Internal Server Error",
				http.StatusInternalServerError)
			log.Println("Profile Update Error:", err)
			return
		}

		// Anonymous/ad-hoc streams are no longer supported. Only reserved
		// profiles (accounts or admin-created profiles) may change settings.
		resolvedProfile, resolveErr := authorization.GetPersonalProfile(token)
		if resolveErr != nil {
			resolvedProfile, resolveErr = authorization.GetPersonalProfileByStreamKey(token)
		}
		if resolveErr != nil || resolvedProfile == nil || !authorization.IsProfileReserved(resolvedProfile.StreamKey) {
			helpers.LogHTTPError(
				responseWriter,
				"This stream key is reserved; use a valid account token",
				http.StatusUnauthorized)
			return
		}

		// Preserve existing title/motd when the client does not send one (e.g. visibility toggle)
		if payload.Motd == "" {
			if existing, err := authorization.GetPersonalProfile(token); err == nil && existing != nil {
				payload.Motd = existing.MOTD
			} else if existing, err := authorization.GetPersonalProfileByStreamKey(token); err == nil && existing != nil {
				payload.Motd = existing.MOTD
			}
		}

		// Try token-based update first (for profile-based streams)
		updatedByToken := false
		err := authorization.UpdateProfile(token, payload.Motd, payload.IsPublic)
		if err == nil {
			updatedByToken = true
		} else {
			// Fallback: update an ad-hoc stream by its public stream key
			err = authorization.UpdateProfileByStreamKey(token, payload.Motd, payload.IsPublic)
			if err != nil {
				helpers.LogHTTPError(
					responseWriter,
					"Internal Server Error",
					http.StatusInternalServerError)
				log.Println(err.Error())
				return
			}
		}

		// Get the updated profile for session update
		var profile *authorization.PersonalProfile
		if updatedByToken {
			profile, _ = authorization.GetPersonalProfile(token)
		} else {
			profile, _ = authorization.GetPersonalProfileByStreamKey(token)
		}
		if profile != nil {
			manager.SessionsManager.UpdateProfile(profile)
		}
	}

	responseWriter.Header().Add("Content-Type", "application/json")
}
