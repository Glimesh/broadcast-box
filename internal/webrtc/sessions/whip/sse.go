package whip

import (
	"encoding/json"
	"log"
)

// Returns all available Video and Audio layers of the provided stream key
func (w *WHIPSession) GetAvailableLayersEvent() string {
	videoLayers := []simulcastLayerResponse{}
	audioLayers := []simulcastLayerResponse{}

	w.TracksLock.RLock()

	// Add available video layers
	for track := range w.VideoTracks {
		videoLayers = append(videoLayers, simulcastLayerResponse{
			EncodingID: w.VideoTracks[track].Rid,
		})
	}

	// Add available audio layers
	for track := range w.AudioTracks {
		audioLayers = append(audioLayers, simulcastLayerResponse{
			EncodingID: w.AudioTracks[track].Rid,
		})
	}

	w.TracksLock.RUnlock()

	resp := map[string]map[string][]simulcastLayerResponse{
		"1": {
			"layers": videoLayers,
		},
		"2": {
			"layers": audioLayers,
		},
	}

	jsonResult, err := json.Marshal(resp)
	if err != nil {
		log.Println("Error converting response", resp, "to Json")
	}

	return "event: layers\ndata: " + string(jsonResult) + "\n\n"
}
