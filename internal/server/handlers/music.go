package handlers

import (
	"encoding/json"
	"net/http"

	adminHandlers "github.com/glimesh/broadcast-box/internal/server/handlers/admin"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

// MusicHandler returns the public site music playlist.
func MusicHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		helpers.LogHTTPError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	playlist, err := adminHandlers.LoadMusicPlaylist()
	if err != nil {
		helpers.LogHTTPError(responseWriter, "Could not load music playlist", http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.Header().Set("Cache-Control", "public, max-age=60")
	_ = json.NewEncoder(responseWriter).Encode(playlist)
}
