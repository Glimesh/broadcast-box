package manager

import (
	"context"
	"log"
	"maps"
	"time"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/session"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whip"
)

// Prepare the Whip Session Manager
func (manager *SessionManager) Setup() {
	log.Println("WhipSessionManager.Setup")

	manager.sessions = make(map[string]*session.Session)
}

// Add new session
func (manager *SessionManager) addSession(profile authorization.PublicProfile) (s *session.Session, err error) {
	log.Println("SessionManager.AddWhipSession")
	activeContext, activeContextCancel := context.WithCancel(context.Background())

	s = &session.Session{

		StreamKey:   profile.StreamKey,
		IsPublic:    profile.IsPublic,
		MOTD:        profile.MOTD,
		StreamStart: time.Now(),

		ActiveContext:       activeContext,
		ActiveContextCancel: activeContextCancel,

		WhepSessions: map[string]*whep.WhepSession{},
	}

	s.HasHost.Store(true)
	manager.sessionsLock.Lock()
	manager.sessions[profile.StreamKey] = s
	manager.sessionsLock.Unlock()

	go s.Snapshot()
	go func() {
		<-activeContext.Done()
		log.Println("SessionManager.Session.Done")

		manager.sessionsLock.Lock()
		delete(manager.sessions, profile.StreamKey)
		manager.sessionsLock.Unlock()

	}()

	return s, nil
}

// Get the stream requested, or create it, and add it to the sessions context
func (manager *SessionManager) GetOrAddSession(profile authorization.PublicProfile, isWhip bool) (session *session.Session, err error) {
	session, ok := manager.GetSessionById(profile.StreamKey)

	if !ok {
		log.Println("SessionManager.GetOrAddStream: Adding", profile.StreamKey)
		session, err = manager.addSession(profile)
	} else if isWhip {
		log.Println("SessionManager.GetOrAddStream: Updating", profile.StreamKey)
		session.UpdateStreamStatus(profile)
	}

	return session, err
}

// Get Session by id
func (manager *SessionManager) GetSessionById(streamKey string) (session *session.Session, foundSession bool) {
	log.Println("SessionManager.GetSessionById", streamKey)

	manager.sessionsLock.RLock()
	defer manager.sessionsLock.RUnlock()

	for _, session := range manager.sessions {
		if streamKey == session.StreamKey {
			return session, true
		}
	}

	return nil, false
}

// Gets the current state of all sessions
func (manager *SessionManager) GetSessionStates(includePrivateStreams bool) (result []session.StreamSessionDto) {
	log.Println("SessionManager.GetSessionStates: IsAdmin", includePrivateStreams)
	manager.sessionsLock.RLock()
	copiedSessions := make(map[string]*session.Session)
	maps.Copy(copiedSessions, manager.sessions)
	manager.sessionsLock.RUnlock()

	for _, s := range copiedSessions {
		s.StatusLock.RLock()

		if !includePrivateStreams && !s.IsPublic {
			s.StatusLock.RUnlock()
			continue
		}

		streamSession := session.StreamSessionDto{
			StreamKey:   s.StreamKey,
			StreamStart: s.StreamStart,
			IsPublic:    s.IsPublic,
			MOTD:        s.MOTD,
			Sessions:    []whep.WhepSessionStateDto{},
			VideoTracks: []session.VideoTrackState{},
			AudioTracks: []session.AudioTrackState{},
		}

		s.StatusLock.RUnlock()

		host := s.Host.Load()
		if host != nil {
			host.TracksLock.RLock()

			for _, audioTrack := range host.AudioTracks {
				streamSession.AudioTracks = append(
					streamSession.AudioTracks,
					session.AudioTrackState{
						Rid:             audioTrack.Rid,
						PacketsReceived: audioTrack.PacketsReceived.Load(),
						PacketsDropped:  audioTrack.PacketsDropped.Load(),
					})
			}

			for _, videoTrack := range host.VideoTracks {
				var lastKeyFrame time.Time
				if value, ok := videoTrack.LastKeyFrame.Load().(time.Time); ok {
					lastKeyFrame = value
				}

				streamSession.VideoTracks = append(
					streamSession.VideoTracks,
					session.VideoTrackState{
						Rid:             videoTrack.Rid,
						Bitrate:         videoTrack.Bitrate.Load(),
						PacketsReceived: videoTrack.PacketsReceived.Load(),
						PacketsDropped:  videoTrack.PacketsDropped.Load(),
						LastKeyframe:    lastKeyFrame,
					})
			}

			host.TracksLock.RUnlock()
		}

		s.WhepSessionsLock.RLock()
		for _, whep := range s.WhepSessions {
			if !whep.IsSessionClosed.Load() {
				streamSession.Sessions = append(streamSession.Sessions, whep.GetWhepSessionStatus())
			}
		}
		s.WhepSessionsLock.RUnlock()

		result = append(result, streamSession)
	}

	return
}

// Update the provided session information
func (manager *SessionManager) UpdateProfile(profile *authorization.PersonalProfile) {
	log.Println("WhipSessionManager.UpdateProfile")
	manager.sessionsLock.RLock()
	whipSession, ok := manager.sessions[profile.StreamKey]
	manager.sessionsLock.RUnlock()

	if ok {
		whipSession.StatusLock.Lock()
		whipSession.MOTD = profile.MOTD
		whipSession.IsPublic = profile.IsPublic
		whipSession.StatusLock.Unlock()
	}
}

// Get Session by id
func (manager *SessionManager) GetWhepSessionById(sessionId string) (whep *whep.WhepSession, foundSession bool) {

	manager.sessionsLock.RLock()
	defer manager.sessionsLock.RUnlock()

	for _, session := range manager.sessions {
		session.WhepSessionsLock.RLock()
		defer session.WhepSessionsLock.RUnlock()
		if whep, ok := session.WhepSessions[sessionId]; ok {
			return whep, true
		}
	}

	return nil, false
}

func (manager *SessionManager) GetHostSessionById(sessionId string) (host *whip.WhipSession, foundSession bool) {
	manager.sessionsLock.RLock()
	defer manager.sessionsLock.RUnlock()

	for _, session := range manager.sessions {
		host := session.Host.Load()
		if host == nil {
			continue
		}

		if sessionId == host.Id {
			return host, true
		}
	}

	return nil, false
}
