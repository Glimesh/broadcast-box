package whip

import (
	"context"
	"log"
	"maps"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/session/whep"
)

// Returns true is no WHIP tracks are present, and no WHEP sessions are waiting for incoming streams
func (whipSession *WhipSession) IsEmpty() bool {
	if whipSession.HasWhepSessions() {
		log.Println("WhipSession.IsEmpty.HasWhepSessions (false):", whipSession.StreamKey)
		return false
	}

	if whipSession.IsActive() {
		log.Println("WhipSession.IsEmpty.IsActive (false):", whipSession.StreamKey)
		return false
	}

	log.Println("WhipSession.IsEmpty (true):", whipSession.StreamKey)
	return true
}

// Returns true if any tracks are available for the session
func (whipSession *WhipSession) IsActive() bool {
	whipSession.TracksLock.RLock()

	if len(whipSession.AudioTracks) != 0 {
		log.Println("WhipSession.IsActive.AudioTracks", len(whipSession.AudioTracks))
		whipSession.TracksLock.RUnlock()
		return true
	}
	if len(whipSession.VideoTracks) != 0 {
		log.Println("WhipSession.IsActive.VideoTracks", len(whipSession.VideoTracks))
		whipSession.TracksLock.RUnlock()
		return true
	}

	whipSession.TracksLock.RUnlock()
	return false
}

func (whipSession *WhipSession) HasWhepSessions() bool {
	log.Println("WhipSession.HasWhepSessions:", len(whipSession.WhepSessions))

	whipSession.WhepSessionsLock.RLock()

	if len(whipSession.WhepSessions) == 0 {
		whipSession.WhepSessionsLock.RUnlock()
		return false
	}

	whipSession.WhepSessionsLock.RUnlock()
	return true
}

func (whipSession *WhipSession) UpdateStreamStatus(profile authorization.PublicProfile) {
	whipSession.StatusLock.Lock()
	whipActiveContext, whipActiveContextCancel := context.WithCancel(context.Background())

	whipSession.HasHost.Store(true)
	whipSession.MOTD = profile.MOTD
	whipSession.IsPublic = profile.IsPublic
	whipSession.ActiveContext = whipActiveContext
	whipSession.ActiveContextCancel = whipActiveContextCancel

	whipSession.StatusLock.Unlock()
}

func (whipSession *WhipSession) GetStreamStatus() (status WhipSessionStatus) {
	whipSession.WhepSessionsLock.RLock()
	whepSessionsCount := len(whipSession.WhepSessions)
	whipSession.WhepSessionsLock.RUnlock()

	whipSession.StatusLock.RLock()

	status = WhipSessionStatus{
		StreamKey:   whipSession.StreamKey,
		MOTD:        whipSession.MOTD,
		ViewerCount: whepSessionsCount,
		IsOnline:    whipSession.HasHost.Load(),
	}

	whipSession.StatusLock.RUnlock()

	return
}
func (whipSession *WhipSession) handleStatus() {
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

func (whipSession *WhipSession) handleAnnounceOffline() {
	// Lock, copy session data, then unlock
	whipSession.WhepSessionsLock.RLock()
	whepSessionsCopy := make(map[string]*whep.WhepSession)
	maps.Copy(whepSessionsCopy, whipSession.WhepSessions)
	whipSession.WhepSessionsLock.RUnlock()

	if len(whepSessionsCopy) == 0 {
		return
	}

	// Generate status
	whipSession.HasHost.Store(false)
	currentStatus := whipSession.GetSessionStatsEvent()

	// Send status to each WHEP session
	for _, whepSession := range whepSessionsCopy {
		if whepSession.IsSessionClosed.Load() {
			continue
		}

		select {
		case whepSession.SseEventsChannel <- currentStatus:
		default:
			log.Println("WhipSession.handleAnnounceOffline: Offline Status update skipped for session (", whepSession.SessionId, ") due to full channel")
		}
	}
}
