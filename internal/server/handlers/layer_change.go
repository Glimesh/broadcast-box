package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
)

type (
	whepLayerRequestJSON struct {
		MediaID    string `json:"mediaId"`
		EncodingID string `json:"encodingId"`
	}
)

func layerChangeHandler(responseWriter http.ResponseWriter, request *http.Request) {
	var requestContent whepLayerRequestJSON

	if err := json.NewDecoder(request.Body).Decode(&requestContent); err != nil {
		helpers.LogHTTPError(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	values := strings.Split(request.URL.RequestURI(), "/")
	whepSessionID := values[len(values)-1]
	whepSession, ok := manager.SessionsManager.GetWHEPSessionByID(whepSessionID)

	slog.Info("Found WHEP session", "sessionID", whepSession.SessionID)

	if !ok {
		helpers.LogHTTPError(responseWriter, "Could not find WHEP session", http.StatusBadRequest)
		return
	}

	if requestContent.MediaID == "1" {
		slog.Info("Setting Video Layer", "encodingID", requestContent.EncodingID)
		whepSession.SetVideoLayer(requestContent.EncodingID)
		return
	}

	if requestContent.MediaID == "2" {
		slog.Info("Setting Audio Layer", "encodingID", requestContent.EncodingID)
		whepSession.SetAudioLayer(requestContent.EncodingID)
		return
	}

	helpers.LogHTTPError(responseWriter, "Unknown media type", http.StatusBadRequest)
}
