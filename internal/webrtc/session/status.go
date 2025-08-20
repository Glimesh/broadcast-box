package session

import (
	"encoding/json"
	"log"
	"maps"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
)

type (
	simulcastLayerResponse struct {
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

func GetSessionStates(whipSessions map[string]*WhipSession, includePrivateStreams bool) []StreamSession {
	sessions := make(map[string]*WhipSession)
	maps.Copy(sessions, whipSessions)

	out := []StreamSession{}

	for streamKey, session := range sessions {
		if !includePrivateStreams && !session.IsPublic {
			continue
		}

		sessionState := StreamSession{
			StreamKey:   streamKey,
			IsPublic:    session.IsPublic,
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

	videoLayers := []simulcastLayerResponse{}
	audioLayers := []simulcastLayerResponse{}

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

func GetSessionStatsJsonString(whipSession *WhipSession) string {
	streamStatus, err := utils.ToJsonString(GetStreamStatus((whipSession)))
	if err != nil {
		log.Println("GetSessionStatsJsonString Error:", err)
		return ""
	}

	return "event: status\ndata: " + streamStatus + "\n\n"
}

func GetWhepSessionStatus(whepSession *WhepSession) string {
	currentAudioLayer := whepSession.AudioLayerCurrent.Load().(string)
	currentVideoLayer := whepSession.VideoLayerCurrent.Load().(string)

	currentSessionState := WhepSessionState{
		Id: whepSession.SessionId,

		AudioLayerCurrent:   currentAudioLayer,
		AudioTimestamp:      whepSession.AudioTimestamp,
		AudioPacketsWritten: whepSession.AudioPacketsWritten,
		AudioSequenceNumber: uint64(whepSession.AudioSequenceNumber),

		VideoLayerCurrent:   currentVideoLayer,
		VideoTimestamp:      whepSession.VideoTimestamp,
		VideoPacketsWritten: whepSession.VideoPacketsWritten,
		VideoSequenceNumber: uint64(whepSession.VideoSequenceNumber),
	}

	currentSessionStateJson, err := utils.ToJsonString(currentSessionState)
	if err != nil {
		log.Println("GetWhepSessionStatus Error:", err)
		return ""
	}

	return "event: currentLayers\ndata: " + currentSessionStateJson + "\n\n"
}
