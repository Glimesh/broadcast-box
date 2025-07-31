package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc"
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

	log.Println("Changing layer", requestContent)
	values := strings.Split(request.URL.RequestURI(), "/")
	whepSessionId := values[len(values)-1]

	if requestContent.MediaId == "1" {
		if err := whepChangeVideoLayer(whepSessionId, requestContent.EncodingId); err != nil {
			helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		}
		return
	}

	if requestContent.MediaId == "2" {
		if err := whepChangeAudioLayer(whepSessionId, requestContent.EncodingId); err != nil {
			helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		}
		return
	}

	helpers.LogHttpError(responseWriter, "Unknown media type", http.StatusBadRequest)
}

func whepChangeAudioLayer(sessionId string, encodingId string) error {
	webrtc.WhipSessionsLock.Lock()
	defer webrtc.WhipSessionsLock.Unlock()

	log.Println("LayerChange.Audio", sessionId)

	for streamKey := range webrtc.WhipSessions {
		webrtc.WhipSessions[streamKey].WhepSessionsLock.Lock()
		defer webrtc.WhipSessions[streamKey].WhepSessionsLock.Unlock()

		if _, ok := webrtc.WhipSessions[streamKey].WhepSessions[sessionId]; ok {
			webrtc.WhipSessions[streamKey].WhepSessions[sessionId].AudioLayerCurrent.Store(encodingId)
			webrtc.WhipSessions[streamKey].PliChan <- true
		}
	}

	log.Println("LayerChange.Audio.Complete", sessionId)
	return nil
}

func whepChangeVideoLayer(sessionId string, encodingId string) error {
	webrtc.WhipSessionsLock.Lock()
	defer webrtc.WhipSessionsLock.Unlock()

	log.Println("LayerChange.Video", sessionId)

	for streamKey := range webrtc.WhipSessions {
		webrtc.WhipSessions[streamKey].WhepSessionsLock.Lock()
		defer webrtc.WhipSessions[streamKey].WhepSessionsLock.Unlock()

		if _, ok := webrtc.WhipSessions[streamKey].WhepSessions[sessionId]; ok {
			webrtc.WhipSessions[streamKey].WhepSessions[sessionId].VideoLayerCurrent.Store(encodingId)
			webrtc.WhipSessions[streamKey].PliChan <- true
		}
	}

	log.Println("LayerChange.Video.Complete", sessionId)
	return nil
}
