package admin

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

// Retrieve all existing profiles
func ProfilesHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if isValidMethod := verifyValidMethod("GET", responseWriter, request); !isValidMethod {
		return
	}

	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHTTPError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	profiles, err := authorization.GetAdminProfilesAll()
	if err != nil {
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(responseWriter).Encode(profiles)
	if err != nil {
		log.Println("API.Admin.Profiles Error", err)
	}
}

type adminTokenResetPayload struct {
	StreamKey string `json:"streamKey"`
}

// Reset the token of an existing stream profile
func ProfilesResetTokenHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if isValidMethod := verifyValidMethod("POST", responseWriter, request); !isValidMethod {
		return
	}

	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHTTPError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	var payload adminTokenResetPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		helpers.LogHTTPError(responseWriter, "Error resolving request", http.StatusBadRequest)
		return
	}

	if err := authorization.ResetProfileToken(payload.StreamKey); err != nil {
		log.Println("API.Admin.ProfilesResetTokenHandler", err)
		helpers.LogHTTPError(responseWriter, "Error updating token", http.StatusBadRequest)
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
}

type adminAddStreamPayload struct {
	StreamKey string `json:"streamKey"`
}

// Reset the token of an existing stream profile
func ProfileAddHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if isValidMethod := verifyValidMethod("POST", responseWriter, request); !isValidMethod {
		return
	}

	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHTTPError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	var payload adminAddStreamPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		helpers.LogHTTPError(responseWriter, "Error resolving request", http.StatusBadRequest)
		return
	}

	if _, err := authorization.CreateProfile(payload.StreamKey); err != nil {
		log.Println("API.Admin.CreateProfile", err)
		helpers.LogHTTPError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
}

type adminRemoveStreamPayload struct {
	StreamKey string `json:"streamKey"`
}

// Reset the token of an existing stream profile
func ProfileRemoveHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if isValidMethod := verifyValidMethod("POST", responseWriter, request); !isValidMethod {
		return
	}

	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHTTPError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	var payload adminRemoveStreamPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		helpers.LogHTTPError(responseWriter, "Error resolving request", http.StatusBadRequest)
		return
	}

	if _, err := authorization.RemoveProfile(payload.StreamKey); err != nil {
		log.Println("API.Admin.RemoveProfile", err)
		helpers.LogHTTPError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
}
