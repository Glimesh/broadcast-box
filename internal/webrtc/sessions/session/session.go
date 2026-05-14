package session

import (
	"fmt"
	"log/slog"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whip"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

func (s *Session) UpdateStreamStatus(profile authorization.PublicProfile) {
	s.StatusLock.Lock()

	s.MOTD = profile.MOTD
	s.IsPublic = profile.IsPublic

	s.StatusLock.Unlock()
}

func (session *Session) SetOnClose(onClose func()) {
	session.onClose = onClose
}

// Add WHEP viewer session
func (s *Session) AddWHEP(whepSessionID string, peerConnection *webrtc.PeerConnection, audioTrack *codecs.TrackMultiCodec, videoTrack *codecs.TrackMultiCodec, videoRTCPSender *webrtc.RTPSender, pliSender func()) (err error) {
	slog.Debug("WHIPSessionManager.WHIPSession.AddWHEPSession")

	whepSession := whep.CreateNewWHEP(
		whepSessionID,
		s.StreamKey,
		audioTrack,
		videoTrack,
		peerConnection,
		pliSender,
		s.ChatManager,
	)

	whepSession.SetOnClose(s.handleWHEPClose)

	s.WHEPSessionsLock.Lock()
	s.WHEPSessions[whepSessionID] = whepSession
	s.WHEPSessionsLock.Unlock()
	s.updateHostWHEPSessionsSnapshot()
	whepSession.RegisterWHEPHandlers(peerConnection)
	go s.handleWHEPVideoRTCPSender(whepSession, videoRTCPSender)

	return nil
}

// Add host
func (s *Session) AddHost(peerConnection *webrtc.PeerConnection) (err error) {
	slog.Debug("Session.AddHost")

	for {
		host := s.Host.Load()
		if host == nil {
			break
		}

		if host.PeerConnection.ConnectionState() != webrtc.PeerConnectionStateClosed {
			return fmt.Errorf("session already has a host")
		}

		if s.Host.CompareAndSwap(host, nil) {
			break
		}
	}

	host := &whip.WHIPSession{
		ID:          uuid.New().String(),
		AudioTracks: make(map[string]*whip.AudioTrack),
		VideoTracks: make(map[string]*whip.VideoTrack),
		ChatManager: s.ChatManager,
	}
	host.SetOnClosed(s.handleHostClosed)

	host.AddPeerConnection(peerConnection, s.StreamKey)
	if !s.Host.CompareAndSwap(nil, host) {
		host.RemovePeerConnection()
		host.RemoveTracks()
		return fmt.Errorf("session already has a host")
	}
	s.resetWHEPSessionsForNewHost()
	host.WHEPSessionsSnapshot.Store(make(map[string]*whep.WHEPSession))
	s.updateHostWHEPSessionsSnapshot()
	s.HasHost.Store(true)

	return nil
}

func (s *Session) RemoveHost() {

	host := s.Host.Swap(nil)
	if host == nil {
		slog.Info("Session.RemoveHost", "streamKey", s.StreamKey, "msg", "No host to remove")
		return
	}

	slog.Info("Session.RemoveHost", "streamKey", s.StreamKey)
	s.HasHost.Store(false)

	host.WHEPSessionsSnapshot.Store(make(map[string]*whep.WHEPSession))
	host.RemovePeerConnection()
	host.RemoveTracks()
}

func (s *Session) handleWHEPClose(whepSessionID string) {
	slog.Info("Session.HandleWHEPClose", "streamKey", s.StreamKey, "whepSessionID", whepSessionID)

	s.WHEPSessionsLock.Lock()
	_, ok := s.WHEPSessions[whepSessionID]
	if ok {
		delete(s.WHEPSessions, whepSessionID)
	}
	s.WHEPSessionsLock.Unlock()

	if !ok {
		return
	}

	s.updateHostWHEPSessionsSnapshot()

	if s.isEmpty() {
		s.close()
	}
}

func (s *Session) handleHostClosed() {
	s.RemoveHost()

	if s.isEmpty() {
		s.close()
	}
}

// Remove all Hosts and clients before closing down session
func (s *Session) close() {
	s.closeOnce.Do(func() {
		s.WHEPSessionsLock.Lock()
		whepSessions := make([]*whep.WHEPSession, 0, len(s.WHEPSessions))
		for _, whepSession := range s.WHEPSessions {
			whepSessions = append(whepSessions, whepSession)
		}
		s.WHEPSessions = make(map[string]*whep.WHEPSession)
		s.WHEPSessionsLock.Unlock()

		for _, whepSession := range whepSessions {
			whepSession.Close()
		}
		s.updateHostWHEPSessionsSnapshot()

		s.RemoveHost()

		if s.onClose != nil {
			s.onClose()
		}
	})
}

func (s *Session) Close() {
	slog.Info("Session.Close", "streamKey", s.StreamKey)
	s.close()
}

// Returns true is no WHIP tracks are present, and no WHEP sessions are waiting for incoming streams
func (s *Session) isEmpty() bool {
	if s.hasWHEPSessions() {
		slog.Debug("Session.IsEmpty.HasWHEPSessions (false)", "streamKey", s.StreamKey)
		return false
	}

	if s.isStreaming() {
		slog.Debug("Session.IsEmpty.IsActive (false)", "streamKey", s.StreamKey)
		return false
	}

	slog.Debug("Session.IsEmpty (true)", "streamKey", s.StreamKey)
	return true
}

// Returns true if any tracks are available for the session
func (s *Session) isStreaming() bool {

	host := s.Host.Load()
	if host == nil {
		return false
	}

	host.TracksLock.RLock()

	if len(host.AudioTracks) != 0 {
		slog.Debug("Session.IsActive.AudioTracks", "count", len(host.AudioTracks))
		host.TracksLock.RUnlock()
		return true
	}
	if len(host.VideoTracks) != 0 {
		slog.Debug("Session.IsActive.VideoTracks", "count", len(host.VideoTracks))
		host.TracksLock.RUnlock()
		return true
	}

	host.TracksLock.RUnlock()
	return false
}

func (s *Session) hasWHEPSessions() bool {
	s.WHEPSessionsLock.RLock()
	slog.Debug("Session.HasWHEPSessions", "count", len(s.WHEPSessions))

	if len(s.WHEPSessions) == 0 {
		s.WHEPSessionsLock.RUnlock()
		return false
	}

	s.WHEPSessionsLock.RUnlock()
	return true
}

func (s *Session) updateHostWHEPSessionsSnapshot() {
	host := s.Host.Load()
	if host == nil {
		return
	}

	s.WHEPSessionsLock.RLock()
	snapshot := make(map[string]*whep.WHEPSession, len(s.WHEPSessions))
	for _, whepSession := range s.WHEPSessions {
		if !whepSession.IsSessionClosed.Load() {
			snapshot[whepSession.SessionID] = whepSession
		}
	}
	s.WHEPSessionsLock.RUnlock()

	host.WHEPSessionsSnapshot.Store(snapshot)
}

func (s *Session) resetWHEPSessionsForNewHost() {
	s.WHEPSessionsLock.RLock()
	for _, whepSession := range s.WHEPSessions {
		if whepSession == nil {
			continue
		}

		whepSession.ResetForNewPublisher()
	}
	s.WHEPSessionsLock.RUnlock()
}

// Get the status of the current session
func (s *Session) GetStreamStatus() (status whipSessionStatus) {
	s.WHEPSessionsLock.RLock()
	whepSessionsCount := len(s.WHEPSessions)
	s.WHEPSessionsLock.RUnlock()

	s.StatusLock.RLock()

	status = whipSessionStatus{
		StreamKey:   s.StreamKey,
		MOTD:        s.MOTD,
		ViewerCount: whepSessionsCount,
		IsOnline:    s.HasHost.Load(),
		StreamStart: s.StreamStart,
	}

	s.StatusLock.RUnlock()

	return
}
