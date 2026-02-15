package whip

import (
	"encoding/json"
	"log"
)

// Returns all available Video and Audio layers of the provided stream key
func (whip *WhipSession) GetAvailableLayersEvent() string {
	videoLayers := []simulcastLayerResponse{}
	audioLayers := []simulcastLayerResponse{}

	whip.TracksLock.RLock()

	// Add available video layers
	for track := range whip.VideoTracks {
		videoLayers = append(videoLayers, simulcastLayerResponse{
			EncodingId: whip.VideoTracks[track].Rid,
		})
	}

	// Add available audio layers
	for track := range whip.AudioTracks {
		audioLayers = append(audioLayers, simulcastLayerResponse{
			EncodingId: whip.AudioTracks[track].Rid,
		})
	}

	whip.TracksLock.RUnlock()

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
