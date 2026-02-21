package manager

import (
	"log"
	"maps"
	"time"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/session"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
)

// Prepare the WHIP Session Manager
func (m *SessionManager) Setup() {
	log.Println("WHIPSessionManager.Setup")

	m.sessions = make(map[string]*session.Session)
}

// Add new session
func (m *SessionManager) addSession(profile authorization.PublicProfile) (s *session.Session, err error) {
	log.Println("SessionManager.AddWHIPSession")

	s = &session.Session{

		StreamKey:   profile.StreamKey,
		IsPublic:    profile.IsPublic,
		MOTD:        profile.MOTD,
		StreamStart: time.Now(),

		WHEPSessions: map[string]*whep.WHEPSession{},
		ChatManager:  m.ChatManager,
	}
	s.SetOnClose(func() {
		log.Println("SessionManager.Session.Done")
		m.sessionsLock.Lock()
		delete(m.sessions, profile.StreamKey)
		m.sessionsLock.Unlock()
	})

	m.sessionsLock.Lock()
	m.sessions[profile.StreamKey] = s
	m.sessionsLock.Unlock()

	return s, nil
}

// Get the stream requested, or create it, and add it to the sessions context
func (m *SessionManager) GetOrAddSession(profile authorization.PublicProfile, isWHIP bool) (session *session.Session, err error) {
	session, ok := m.GetSessionByID(profile.StreamKey)

	if !ok {
		log.Println("SessionManager.GetOrAddStream: Adding", profile.StreamKey)
		session, err = m.addSession(profile)
	} else if isWHIP {
		log.Println("SessionManager.GetOrAddStream: Updating", profile.StreamKey)
		session.UpdateStreamStatus(profile)
	}

	return session, err
}

// Get Session by id
func (m *SessionManager) GetSessionByID(streamKey string) (session *session.Session, foundSession bool) {
	log.Println("SessionManager.GetSessionByID", streamKey)

	m.sessionsLock.RLock()
	defer m.sessionsLock.RUnlock()

	session, foundSession = m.sessions[streamKey]
	return session, foundSession
}

// Gets the current state of all sessions
func (m *SessionManager) GetSessionStates(includePrivateStreams bool) (result []session.StreamSessionState) {
	log.Println("SessionManager.GetSessionStates: IsAdmin", includePrivateStreams)
	m.sessionsLock.RLock()
	copiedSessions := make(map[string]*session.Session)
	maps.Copy(copiedSessions, m.sessions)
	m.sessionsLock.RUnlock()

	for _, s := range copiedSessions {
		s.StatusLock.RLock()

		if !includePrivateStreams && !s.IsPublic {
			s.StatusLock.RUnlock()
			continue
		}

		streamSession := session.StreamSessionState{
			StreamKey:   s.StreamKey,
			StreamStart: s.StreamStart,
			IsPublic:    s.IsPublic,
			MOTD:        s.MOTD,
			Sessions:    []whep.SessionState{},
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

		s.WHEPSessionsLock.RLock()
		for _, whep := range s.WHEPSessions {
			if !whep.IsSessionClosed.Load() {
				streamSession.Sessions = append(streamSession.Sessions, whep.GetWHEPSessionStatus())
			}
		}
		s.WHEPSessionsLock.RUnlock()

		result = append(result, streamSession)
	}

	return
}

// Update the provided session information
func (m *SessionManager) UpdateProfile(profile *authorization.PersonalProfile) {
	log.Println("WHIPSessionManager.UpdateProfile")
	m.sessionsLock.RLock()
	whipSession, ok := m.sessions[profile.StreamKey]
	m.sessionsLock.RUnlock()

	if ok {
		whipSession.StatusLock.Lock()
		whipSession.MOTD = profile.MOTD
		whipSession.IsPublic = profile.IsPublic
		whipSession.StatusLock.Unlock()
	}
}

// Get Session by id
func (m *SessionManager) GetWHEPSessionByID(sessionID string) (whep *whep.WHEPSession, foundSession bool) {
	_, whepSession, foundSession := m.GetSessionAndWHEPByID(sessionID)
	return whepSession, foundSession
}

func (m *SessionManager) SendPLIByWHEPSessionID(sessionID string) {
	streamSession, _, foundSession := m.GetSessionAndWHEPByID(sessionID)
	if !foundSession {
		log.Println("SessionManager.SendPLIByWHEPSessionID: WHEP session not found", sessionID)
		return
	}

	host := streamSession.Host.Load()
	if host == nil {
		log.Println(
			"SessionManager.SendPLIByWHEPSessionID: WHIP session not found",
			"whepSessionID", sessionID,
			"streamKey", streamSession.StreamKey,
		)
		return
	}

	host.SendPLI()
}

func (m *SessionManager) GetSessionAndWHEPByID(sessionID string) (streamSession *session.Session, whepSession *whep.WHEPSession, foundSession bool) {
	m.sessionsLock.RLock()
	defer m.sessionsLock.RUnlock()

	for _, session := range m.sessions {
		session.WHEPSessionsLock.RLock()
		whepSession, ok := session.WHEPSessions[sessionID]
		session.WHEPSessionsLock.RUnlock()
		if ok {
			return session, whepSession, true
		}
	}

	return nil, nil, false
}

func (m *SessionManager) GetSessionByHostSessionID(sessionID string) (session *session.Session, foundSession bool) {
	m.sessionsLock.RLock()
	defer m.sessionsLock.RUnlock()

	for _, session := range m.sessions {
		host := session.Host.Load()
		if host == nil {
			continue
		}

		if sessionID == host.ID {
			return session, true
		}
	}

	return nil, false
}
