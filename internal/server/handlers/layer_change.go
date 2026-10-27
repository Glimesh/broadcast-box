package handlers

import (
	"encoding/json"
	"log"
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

	log.Println("Found WHEP session", whepSession.SessionID)

	if !ok {
		helpers.LogHTTPError(responseWriter, "Could not find WHEP session", http.StatusBadRequest)
		return
	}

	if requestContent.MediaID == "1" {
		log.Println("Setting Video Layer", requestContent.EncodingID)
		whepSession.SetVideoLayer(requestContent.EncodingID)
		return
	}

	if requestContent.MediaID == "2" {
		log.Println("Setting Audio Layer", requestContent.EncodingID)
		whepSession.SetAudioLayer(requestContent.EncodingID)
		return
	}

	helpers.LogHTTPError(responseWriter, "Unknown media type", http.StatusBadRequest)
}
