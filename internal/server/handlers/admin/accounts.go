package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	accountStore "github.com/glimesh/broadcast-box/internal/accounts"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

type adminAccount struct {
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	StreamKey       string `json:"streamKey"`
	CanUploadEmotes bool   `json:"canUploadEmotes"`
	CreatedAt       string `json:"createdAt"`
}

// AccountsHandler returns all registered accounts.
func AccountsHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if isValidMethod := verifyValidMethod("GET", responseWriter, request); !isValidMethod {
		return
	}

	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHTTPError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	accounts, err := accountStore.List()
	if err != nil {
		helpers.LogHTTPError(responseWriter, "Could not load accounts", http.StatusInternalServerError)
		return
	}

	result := []adminAccount{}
	for _, account := range accounts {
		result = append(result, adminAccount{
			Username:        account.Username,
			DisplayName:     account.DisplayName,
			StreamKey:       account.StreamKey,
			CanUploadEmotes: account.CanUploadEmotes,
			CreatedAt:       account.CreatedAt.Format(time.RFC3339),
		})
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(responseWriter).Encode(result); err != nil {
		log.Println("API.Admin.Accounts Error", err)
	}
}

type adminDeleteAccountsPayload struct {
	Usernames []string `json:"usernames"`
}

// AccountsDeleteHandler deletes one or more accounts.
func AccountsDeleteHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if isValidMethod := verifyValidMethod("POST", responseWriter, request); !isValidMethod {
		return
	}

	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHTTPError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	var payload adminDeleteAccountsPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		helpers.LogHTTPError(responseWriter, "Error resolving request", http.StatusBadRequest)
		return
	}

	deleted := []string{}
	failed := map[string]string{}

	for _, username := range payload.Usernames {
		if err := accountStore.Delete(username); err != nil {
			failed[username] = err.Error()
			continue
		}

		deleted = append(deleted, username)
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(responseWriter).Encode(map[string]any{
		"deleted": deleted,
		"failed":  failed,
	}); err != nil {
		log.Println("API.Admin.AccountsDelete Error", err)
	}
}

type adminAccountPermissionPayload struct {
	Username        string `json:"username"`
	CanUploadEmotes bool   `json:"canUploadEmotes"`
}

// AccountPermissionHandler updates whether an account may upload custom emotes.
func AccountPermissionHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if isValidMethod := verifyValidMethod("POST", responseWriter, request); !isValidMethod {
		return
	}

	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHTTPError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	var payload adminAccountPermissionPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		helpers.LogHTTPError(responseWriter, "Error resolving request", http.StatusBadRequest)
		return
	}

	if err := accountStore.SetPermission(payload.Username, payload.CanUploadEmotes); err != nil {
		helpers.LogHTTPError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.WriteHeader(http.StatusOK)
}
