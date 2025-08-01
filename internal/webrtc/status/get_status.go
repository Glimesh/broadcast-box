package status

import (
	"log"
	"maps"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc"
	"github.com/glimesh/broadcast-box/internal/webrtc/stream"
)

func GetStreamState() StreamState {
	// TODO Implement
	return StreamState{}
}

func GetStreamStates() []StreamState {
	webrtc.WhipSessionsLock.Lock()
	defer webrtc.WhipSessionsLock.Unlock()
	sessions := make(map[string]*stream.WhipSession)
	maps.Copy(sessions, webrtc.WhipSessions)

	out := []StreamState{}

	for streamKey, session := range sessions {
		if !session.IsPublic {
			continue
		}

		sessionState := StreamState{
			StreamKey:   streamKey,
			Sessions:    []WhepSessionState{},
			VideoTracks: []VideoTrackState{},
			AudioTracks: []AudioTrackState{},
		}

		for id, whepSession := range webrtc.WhipSessions[streamKey].WhepSessions {
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
