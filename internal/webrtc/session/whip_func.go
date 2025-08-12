package session

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
)

// TODO:
// When 2 different machines start a session, the latter will overwrite the session id
// and take over the session. To allow for a smooth multiuser stream against one streamKey,
// make the session Id into a list of session Ids that are maintained as sessions are added and removed.
// Doing so will also allow for layer changes to be connected, so that changing the video feed triggers the
// corresponding audio feed as well, if at another session
func GetStream(profile authorization.Profile, whipSessionId string) (*WhipSession, error) {
	WhipSessionsLock.Lock()
	stream, ok := WhipSessions[profile.StreamKey]

	if !ok {
		whipActiveContext, whipActiveContextCancel := context.WithCancel(context.Background())

		stream = &WhipSession{
			SessionId: whipSessionId,

			StreamKey: strings.ReplaceAll(profile.StreamKey, " ", ""),
			IsPublic:  profile.IsPublic,
			MOTD:      profile.MOTD,

			ActiveContext:       whipActiveContext,
			ActiveContextCancel: whipActiveContextCancel,
			PliChan:             make(chan any, 250),
			OnOnlineChan:        make(chan bool, 5),
			OnTrackChan:         make(chan struct{}, 5),
			SSEChan:             make(chan any, 5),

			AudioTracks: []*AudioTrack{},
			VideoTracks: []*VideoTrack{},

			WhepSessions: map[string]*WhepSession{},
		}

		WhipSessions[profile.StreamKey] = stream
	}

	if whipSessionId != "" {
		stream.SessionId = whipSessionId
		stream.HasHost.Store(true)
		stream.MOTD = profile.MOTD
	}

	WhipSessionsLock.Unlock()
	return stream, nil
}

func StartWhipSessionLoop(stream *WhipSession) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		// All WHIP and WHEP sessions has left, close down
		case <-stream.ActiveContext.Done():
			ticker.Stop()
			return

		// Announce new layers available
		case <-stream.OnTrackChan:
			// Lock, copy session data, then unlock
			WhipSessionsLock.Lock()
			stream.WhepSessionsLock.RLock()

			whepSessionsCopy := make([]*WhepSession, 0, len(stream.WhepSessions))
			for _, whep := range stream.WhepSessions {
				whepSessionsCopy = append(whepSessionsCopy, whep)
			}

			stream.WhepSessionsLock.RUnlock()
			WhipSessionsLock.Unlock()

			// Generate layer info outside lock
			currentLayers := GetAvailableLayersJsonString(stream)

			// Send to each WHEP session
			for _, whep := range whepSessionsCopy {
				select {
				case whep.SSEChannel <- currentLayers:
				default:
					log.Println("WHIP.Loop: OnTrackChange - Channel full, skipping update")
				}
			}

		// Send status every 5 seconds
		case <-ticker.C:
			// Lock, copy session data, then unlock
			stream.WhepSessionsLock.RLock()
			whepSessionsCopy := make([]*WhepSession, 0, len(stream.WhepSessions))
			for _, whep := range stream.WhepSessions {
				whepSessionsCopy = append(whepSessionsCopy, whep)
			}
			stream.WhepSessionsLock.RUnlock()

			// Generate status
			currentStatus := GetSessionStatsJsonString(stream)

			// Send to WHIP session
			stream.SSEChan <- currentStatus

			// Send to each WHEP session
			for _, whep := range whepSessionsCopy {
				select {
				case whep.SSEChannel <- currentStatus:
				default:
					log.Println("WHIP.Loop: Status update skipped for session due to full channel")
				}
			}
		}
	}
}
