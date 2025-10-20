package whip

import (
	"context"
	"log"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
)

// Returns true is no WHIP streams are present, and not WHEP sessions are waiting for incoming streams
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
	whipSession.WhepSessionsLock.RLock()

	log.Println("WhipSession.HasWhepSessions:", len(whipSession.WhepSessions))

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
