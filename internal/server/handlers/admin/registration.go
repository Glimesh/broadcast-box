package admin

import (
	"encoding/json"
	"net/http"

	accountStore "github.com/glimesh/broadcast-box/internal/accounts"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

type registrationStatus struct {
	Enabled bool `json:"enabled"`
}

// RegistrationHandler reports or changes whether new accounts can be created.
func RegistrationHandler(responseWriter http.ResponseWriter, request *http.Request) {
	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHTTPError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	switch request.Method {
	case http.MethodGet:
		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(registrationStatus{
			Enabled: accountStore.RegistrationEnabled(),
		})

	case http.MethodPost:
		var payload struct {
			Enabled bool `json:"enabled"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			helpers.LogHTTPError(responseWriter, "Invalid payload", http.StatusBadRequest)
			return
		}

		if err := accountStore.SetRegistrationEnabled(payload.Enabled); err != nil {
			helpers.LogHTTPError(responseWriter, "Could not update registration", http.StatusInternalServerError)
			return
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(registrationStatus{
			Enabled: payload.Enabled,
		})

	default:
		helpers.LogHTTPError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
