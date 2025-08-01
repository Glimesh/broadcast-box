package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc"
)

func sseHandler(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("SSEHandler Called")
	responseWriter.Header().Add("Content-Type", "text/event-stream")
	responseWriter.Header().Add("Cache-Control", "no-cache")
	responseWriter.Header().Add("Connection", "keep-alive")

	values := strings.Split(request.URL.RequestURI(), "/")
	whepSessionId := values[len(values)-1]

	layers, err := layers(whepSessionId)
	if err != nil {
		helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err = fmt.Fprintf(responseWriter, "event: layers\ndata: %s\n\n\n", string(layers)); err != nil {
		log.Println(err)
	}
}

func layers(whepSessionId string) ([]byte, error) {
	webrtc.WhipSessionsLock.Lock()
	defer webrtc.WhipSessionsLock.Unlock()

	videoLayers := []simulcaseLayerResponse{}
	audioLayers := []simulcaseLayerResponse{}
	for streamKey := range webrtc.WhipSessions {
		webrtc.WhipSessions[streamKey].WhepSessionsLock.Lock()
		defer webrtc.WhipSessions[streamKey].WhepSessionsLock.Unlock()

		if _, ok := webrtc.WhipSessions[streamKey].WhepSessions[whepSessionId]; ok {

			for track := range webrtc.WhipSessions[streamKey].VideoTracks {
				videoLayers = append(videoLayers, simulcaseLayerResponse{
					EncodingId: webrtc.WhipSessions[streamKey].VideoTracks[track].Rid,
				})
			}

			for track := range webrtc.WhipSessions[streamKey].AudioTracks {
				audioLayers = append(audioLayers, simulcaseLayerResponse{
					EncodingId: webrtc.WhipSessions[streamKey].AudioTracks[track].Rid,
				})
			}
		}
	}

	resp := map[string]map[string][]simulcaseLayerResponse{
		"1": {
			"layers": videoLayers,
		},
		"2": {
			"layers": audioLayers,
		},
	}

	return json.Marshal(resp)
}

type (
	simulcaseLayerResponse struct {
		EncodingId string `json:"encodingId"`
	}
)
