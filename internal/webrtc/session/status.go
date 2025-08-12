package session

import (
	"encoding/json"
	"log"
	"maps"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
)

type (
	simulcaseLayerResponse struct {
		EncodingId string `json:"encodingId"`
	}
)

func GetStreamStatus(session *WhipSession) (status StreamStatus) {
	session.StatusLock.Lock()
	defer session.StatusLock.Unlock()

	return StreamStatus{
		StreamKey:   session.StreamKey,
		MOTD:        session.MOTD,
		ViewerCount: len(session.WhepSessions),
		IsOnline:    session.HasHost.Load(),
	}
}

func GetSessionStates(whipSessions map[string]*WhipSession) []StreamSession {
	sessions := make(map[string]*WhipSession)
	maps.Copy(sessions, whipSessions)

	out := []StreamSession{}

	for streamKey, session := range sessions {
		if !session.IsPublic {
			continue
		}

		sessionState := StreamSession{
			StreamKey:   streamKey,
			Sessions:    []WhepSessionState{},
			VideoTracks: []VideoTrackState{},
			AudioTracks: []AudioTrackState{},
		}

		for id, whepSession := range whipSessions[streamKey].WhepSessions {
			currentAudioLayer, ok := whepSession.AudioLayerCurrent.Load().(string)
			if !ok {
				log.Println("GetStates", id, "could not find an active audio layer")
				continue
			}
			currentVideoLayer, ok := whepSession.VideoLayerCurrent.Load().(string)
			if !ok {
				log.Println("GetStates", id, "could not find an active audio layer")
				continue
			}

			sessionState.Sessions = append(
				sessionState.Sessions,
				WhepSessionState{
					Id: id,

					AudioLayerCurrent:   currentAudioLayer,
					AudioTimestamp:      whepSession.AudioTimestamp,
					AudioPacketsWritten: whepSession.AudioPacketsWritten,
					AudioSequenceNumber: uint64(whepSession.AudioSequenceNumber),

					VideoLayerCurrent:   currentVideoLayer,
					VideoTimestamp:      whepSession.VideoTimestamp,
					VideoPacketsWritten: whepSession.VideoPacketsWritten,
					VideoSequenceNumber: uint64(whepSession.VideoSequenceNumber),
				})
		}

		for _, audioTrack := range session.AudioTracks {
			sessionState.AudioTracks = append(
				sessionState.AudioTracks,
				AudioTrackState{
					Rid:             audioTrack.Rid,
					PacketsReceived: audioTrack.PacketsReceived.Load(),
				})
		}

		for _, videoTrack := range session.VideoTracks {
			var lastKeyFrame time.Time
			if value, ok := videoTrack.LastKeyFrame.Load().(time.Time); ok {
				lastKeyFrame = value
			}

			sessionState.VideoTracks = append(
				sessionState.VideoTracks,
				VideoTrackState{
					Rid:             videoTrack.Rid,
					PacketsReceived: videoTrack.PacketsReceived.Load(),
					LastKeyframe:    lastKeyFrame,
				})
		}

		out = append(out, sessionState)
	}

	return out
}

// Returns all available Video and Audio layers of the provided stream key
func GetAvailableLayersJsonString(whipSession *WhipSession) string {
	whipSession.TracksLock.RLock()
	defer whipSession.TracksLock.RUnlock()

	videoLayers := []simulcaseLayerResponse{}
	audioLayers := []simulcaseLayerResponse{}

	// Add available video layers
	for track := range whipSession.VideoTracks {
		videoLayers = append(videoLayers, simulcaseLayerResponse{
			EncodingId: whipSession.VideoTracks[track].Rid,
		})
	}

	// Add available audio layers
	for track := range whipSession.AudioTracks {
		audioLayers = append(audioLayers, simulcaseLayerResponse{
			EncodingId: whipSession.AudioTracks[track].Rid,
		})
	}

	resp := map[string]map[string][]simulcaseLayerResponse{
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

func GetSessionStatsJsonString(whipSession *WhipSession) string {
	return "event: status\ndata: " + utils.ToJsonString(GetStreamStatus(whipSession)) + "\n\n"
}
