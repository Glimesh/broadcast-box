package whip

import (
	"encoding/json"
	"log"
	"maps"

	"github.com/glimesh/broadcast-box/internal/webrtc/session/whep"
	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
)

// Send out an event to all WHEP sessions to notify that available layers has changed
func (whipSession *WhipSession) AnnounceLayersToWhepClients() {
	log.Println("WhipSession.Loop: Announce Layers to clients", whipSession.StreamKey)

	// Lock, copy session data, then unlock
	whipSession.WhepSessionsLock.RLock()
	whepSessionsCopy := make(map[string]*whep.WhepSession)
	maps.Copy(whepSessionsCopy, whipSession.WhepSessions)
	whipSession.WhepSessionsLock.RUnlock()

	// Generate layer info outside lock
	currentLayers := whipSession.GetAvailableLayersEvent()

	// Send to each WHEP session
	for _, whepSession := range whepSessionsCopy {
		if !whepSession.IsSessionClosed.Load() {
			select {
			case whepSession.SseEventsChannel <- currentLayers:
			default:
				log.Println("WhipSession.AnnounceLayersToWhepClients: Channel full, skipping update (SessionId:", whepSession.SessionId, ")")
			}
		}
	}
}

// Send out an event to all WHEP sessions to notify that available layers has changed
func (whipSession *WhipSession) AnnounceStreamStartToWhepClients() {
	log.Println("WhipSession.AnnounceStreamStartToWhepClients:", whipSession.StreamKey)

	// Lock, copy session data, then unlock
	whipSession.WhepSessionsLock.RLock()
	whepSessionsCopy := make(map[string]*whep.WhepSession)
	maps.Copy(whepSessionsCopy, whipSession.WhepSessions)
	whipSession.WhepSessionsLock.RUnlock()

	// Generate layer info outside lock
	streamStartMessage := "event: streamStart\ndata:\n"

	// Send to each WHEP session
	for _, whepSession := range whepSessionsCopy {
		if !whepSession.IsSessionClosed.Load() {

			// Announce to frontend
			select {
			case whepSession.SseEventsChannel <- streamStartMessage:
			default:
				log.Println("WhepSession.AnnounceStreamStartToWhepClients: Channel full, skipping update (SessionId:", whepSession.SessionId, ")")
			}

			// Announce internally
			select {
			case whepSession.WhipEventsChannel <- "active":
			default:
				log.Println("WhepSession.WhipEventsChannel: Channel full, skipping update (SessionId:", whepSession.SessionId, ")")
			}
		}
	}
}

func (whipSession *WhipSession) GetSessionStatsEvent() string {
	//TODO: WhepSessionsSnapshot should only contain information about the current state of the session, not
	// references to chans and other types that cannot be json serialized

	// status, err := utils.ToJsonString(whipSession.WhepSessionsSnapshot.Load().(map[string]*whep.WhepSession))
	status, err := utils.ToJsonString(whipSession.GetStreamStatus())
	if err != nil {
		log.Println("GetSessionStatsJsonString Error:", err)
		return ""
	}

	return "event: status\ndata: " + status + "\n\n"
}

// Returns all available Video and Audio layers of the provided stream key
func (whipSession *WhipSession) GetAvailableLayersEvent() string {
	videoLayers := []simulcastLayerResponse{}
	audioLayers := []simulcastLayerResponse{}

	whipSession.TracksLock.RLock()

	// Add available video layers
	for track := range whipSession.VideoTracks {
		videoLayers = append(videoLayers, simulcastLayerResponse{
			EncodingId: whipSession.VideoTracks[track].Rid,
		})
	}

	// Add available audio layers
	for track := range whipSession.AudioTracks {
		audioLayers = append(audioLayers, simulcastLayerResponse{
			EncodingId: whipSession.AudioTracks[track].Rid,
		})
	}

	whipSession.TracksLock.RUnlock()

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
