package whip

import (
	"log"
	"maps"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/session/whep"
)

func (whipSession *WhipSession) StartWhipSessionStatusLoop() {
	log.Println("WhipSession.StartWhipSessionStatusLoop")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	whipSession.AnnounceStreamStartToWhepClients()

	for {
		select {

		// Whip session is shutting down
		case <-whipSession.ActiveContext.Done():
			log.Println("WhipSession.StartWhipSessionStatusLoop.Done")
			return

		// Announce new layers available
		case <-whipSession.OnTrackChangeChannel:
			log.Println("WhipSession.AnnounceLayersToWhepClients")
			whipSession.AnnounceLayersToWhepClients()
			log.Println("WhipSession.AnnounceLayersToWhepClients.Done")

		// Send status every 5 seconds
		case <-ticker.C:
			//TODO: Make this more event based so that a 5 second trigger is not needed
			whipSession.statusTick()
		}
	}
}

func (whipSession *WhipSession) statusTick() {
	// Lock, copy session data, then unlock
	whipSession.WhepSessionsLock.RLock()
	whepSessionsCopy := make(map[string]*whep.WhepSession)
	maps.Copy(whepSessionsCopy, whipSession.WhepSessions)
	whipSession.WhepSessionsLock.RUnlock()

	whipSession.TracksLock.RLock()
	videoTrackCount := len(whipSession.VideoTracks)
	audioTrackCount := len(whipSession.AudioTracks)
	whipSession.TracksLock.RUnlock()

	hasActiveHost := videoTrackCount != 0 || audioTrackCount != 0
	if hasActiveHost {
		whipSession.HasHost.Store(true)
	} else {
		whipSession.HasHost.Store(false)
		whipSession.ActiveContextCancel()
	}

	if len(whepSessionsCopy) == 0 {
		return
	}

	// Generate status
	currentStatus := whipSession.GetSessionStatsEvent()

	// Send status to each WHEP session
	for _, whepSession := range whepSessionsCopy {
		if whepSession.IsSessionClosed.Load() {
			continue
		}

		select {
		case whepSession.SseEventsChannel <- currentStatus:
		default:
			log.Println("WhipSession.Loop.StatusTick: Status update skipped for session (", whepSession.SessionId, ") due to full channel")
		}
	}
}

// Get highest prioritized audio track in the whip session
// This only works if the priority has been set.
// Currently this is only supported when being set through the simulcast
// property in the offer made by the whip connection
func (whipSession *WhipSession) GetHighestPrioritizedAudioTrack() string {
	if len(whipSession.AudioTracks) != 0 {
		highestPriorityAudioTrack := whipSession.AudioTracks[0]
		for _, trackPriority := range whipSession.AudioTracks[1:] {
			if trackPriority.Priority < highestPriorityAudioTrack.Priority {
				highestPriorityAudioTrack = trackPriority
			}
		}

		return highestPriorityAudioTrack.Rid
	}

	return ""
}

// Get highest prioritized video track in the whip session
// This only works if the priority has been set.
// Currently this is only supported when being set through the simulcast
// property in the offer made by the whip connection
func (whipSession *WhipSession) GetHighestPrioritizedVideoTrack() string {
	if len(whipSession.VideoTracks) != 0 {
		highestPriorityVideoTrack := whipSession.VideoTracks[0]
		for _, trackPriority := range whipSession.VideoTracks[1:] {
			if trackPriority.Priority < highestPriorityVideoTrack.Priority {
				highestPriorityVideoTrack = trackPriority
			}
		}

		return highestPriorityVideoTrack.Rid
	}

	return ""
}
