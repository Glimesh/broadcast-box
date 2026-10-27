package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	accountStore "github.com/glimesh/broadcast-box/internal/accounts"
	accountsHandlers "github.com/glimesh/broadcast-box/internal/server/handlers/accounts"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
)

type chatDeletePayload struct {
	StreamKey string `json:"streamKey"`
	MessageID string `json:"messageId"`
}

// ChatDeleteHandler lets an account owner or moderator delete a chat message.
func ChatDeleteHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		helpers.LogHTTPError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	account, _, ok := accountsHandlers.CurrentAccount(request)
	if !ok {
		helpers.LogHTTPError(responseWriter, "Not signed in", http.StatusUnauthorized)
		return
	}

	var payload chatDeletePayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		helpers.LogHTTPError(responseWriter, "Invalid payload", http.StatusBadRequest)
		return
	}

	streamKey := strings.TrimSpace(payload.StreamKey)
	messageID := strings.TrimSpace(payload.MessageID)
	if streamKey == "" || messageID == "" {
		helpers.LogHTTPError(responseWriter, "Missing stream key or message id", http.StatusBadRequest)
		return
	}

	owner, err := accountStore.Get(streamKey)
	if err != nil {
		helpers.LogHTTPError(responseWriter, "Stream not found", http.StatusNotFound)
		return
	}

	allowed := strings.EqualFold(account.DisplayName, owner.DisplayName)
	if !allowed {
		for _, mod := range owner.ChatMods {
			if strings.EqualFold(mod, account.DisplayName) {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		helpers.LogHTTPError(responseWriter, "Not authorized to delete messages", http.StatusForbidden)
		return
	}

	if manager.SessionsManager == nil || manager.SessionsManager.ChatManager == nil {
		helpers.LogHTTPError(responseWriter, "Chat is not available", http.StatusServiceUnavailable)
		return
	}

	if err := manager.SessionsManager.ChatManager.DeleteMessage(streamKey, messageID, account.DisplayName); err != nil {
		helpers.LogHTTPError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}
