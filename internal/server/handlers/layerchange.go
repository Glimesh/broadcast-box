package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
)

type (
	whepLayerRequestJson struct {
		MediaId    string `json:"mediaId"`
		EncodingId string `json:"encodingId"`
	}
)

func layerChangeHandler(responseWriter http.ResponseWriter, request *http.Request) {
	var requestContent whepLayerRequestJson

	if err := json.NewDecoder(request.Body).Decode(&requestContent); err != nil {
		helpers.LogHttpError(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	values := strings.Split(request.URL.RequestURI(), "/")
	whepSessionId := values[len(values)-1]
	whepSession, ok := session.SessionManager.GetWhepStream(whepSessionId)

	log.Println("Found WHEP session", whepSession.SessionId)

	if !ok {
		helpers.LogHttpError(responseWriter, "Could not find WHEP session", http.StatusBadRequest)
		return
	}

	if requestContent.MediaId == "1" {
		log.Println("Setting Video Layer", requestContent.EncodingId)
		whepSession.SetVideoLayer(requestContent.EncodingId)
		return
	}

	if requestContent.MediaId == "2" {
		log.Println("Setting Audio Layer", requestContent.EncodingId)
		whepSession.SetAudioLayer(requestContent.EncodingId)
		return
	}

	helpers.LogHttpError(responseWriter, "Unknown media type", http.StatusBadRequest)
}
