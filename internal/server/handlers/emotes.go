package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	accountStore "github.com/glimesh/broadcast-box/internal/accounts"
	accountsHandlers "github.com/glimesh/broadcast-box/internal/server/handlers/accounts"
	adminHandlers "github.com/glimesh/broadcast-box/internal/server/handlers/admin"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

// EmotesHandler returns the public list of custom emotes.
func EmotesHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		helpers.LogHTTPError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	emotes, err := adminHandlers.ListEmotes()
	if err != nil {
		helpers.LogHTTPError(responseWriter, "Could not list emotes", http.StatusInternalServerError)
		return
	}

	// Add stream-owner emotes when the stream key belongs to a permitted account.
	streamKey := request.URL.Query().Get("streamKey")
	if streamKey != "" {
		if account, err := accountStore.Get(streamKey); err == nil && account.CanUploadEmotes {
			userEmotes, err := accountsHandlers.ListUserEmotes(account.Username)
			if err == nil {
				seen := map[string]bool{}
				for _, emote := range emotes {
					seen[strings.ToLower(emote.Name)] = true
				}
				for _, emote := range userEmotes {
					if !seen[strings.ToLower(emote.Name)] {
						emotes = append(emotes, emote)
						seen[strings.ToLower(emote.Name)] = true
					}
				}
			}
		}
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.Header().Set("Cache-Control", "public, max-age=60")
	if err := json.NewEncoder(responseWriter).Encode(emotes); err != nil {
		log.Println("API.Emotes Error:", err)
	}
}
