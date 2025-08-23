package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

func adminProfilesHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if isValidMethod := verifyValidMethod("GET", responseWriter, request); isValidMethod == false {
		return
	}

	sessionResult := verifyAdminSession(request)
	if sessionResult.IsValid != true {
		helpers.LogHttpError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	profiles, err := authorization.GetAdminProfilesAll()
	if err != nil {
		helpers.LogHttpError(responseWriter, "Error loading profiles", http.StatusBadRequest)
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	json.NewEncoder(responseWriter).Encode(profiles)
}

func adminProfilesResetTokenHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if isValidMethod := verifyValidMethod("POST", responseWriter, request); isValidMethod == false {
		return
	}

	sessionResult := verifyAdminSession(request)
	if sessionResult.IsValid != true {
		helpers.LogHttpError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	var payload adminTokenResetPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		helpers.LogHttpError(responseWriter, "Error resolving request", http.StatusBadRequest)
		return
	}

	if err := authorization.ResetProfileToken(payload.StreamKey); err != nil {
		log.Println("WPI.Admin.ProfilesResetTokenHandler", err)
		helpers.LogHttpError(responseWriter, "Error updating token", http.StatusBadRequest)
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
}

type adminTokenResetPayload struct {
	StreamKey string `json:"streamKey"`
}
