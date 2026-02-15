package session

import (
	"context"
	"fmt"
	"log"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whip"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

// Get Whip stream by stream key
func (session *Session) GetHost(streamKey string) (host *whip.WhipSession, foundSession bool) {
	log.Println("Session.GetHost")

	host = session.Host.Load()
	if host == nil {
		return nil, false
	}

	return host, true
}

// Find Whep session by session id
func (session *Session) GetWhepStream(sessionId string) (whepSession *whep.WhepSession, foundSession bool) {
	log.Println("WhipSessionManager.GetWhepStream")

	session.WhepSessionsLock.RLock()
	defer session.WhepSessionsLock.RUnlock()

	if whepSession, ok := session.WhepSessions[sessionId]; ok {
		return whepSession, true
	}

	return nil, false
}

func (session *Session) UpdateStreamStatus(profile authorization.PublicProfile) {
	session.StatusLock.Lock()

	session.HasHost.Store(true)
	session.MOTD = profile.MOTD
	session.IsPublic = profile.IsPublic

	session.StatusLock.Unlock()
}

// Add WHEP session to existing WHIP session
func (session *Session) AddWhep(whepSessionId string, peerConnection *webrtc.PeerConnection, audioTrack *codecs.TrackMultiCodec, videoTrack *codecs.TrackMultiCodec, videoRtcpSender *webrtc.RTPSender) (err error) {
	log.Println("WhipSessionManager.WhipSession.AddWhepSession")

	host := session.Host.Load()
	if host == nil {
		return fmt.Errorf("no host was found on the current session")
	}

	whepSession := whep.CreateNewWhep(
		whepSessionId,
		audioTrack,
		host.GetHighestPrioritizedAudioTrack(),
		videoTrack,
		host.GetHighestPrioritizedVideoTrack(),
		peerConnection)

	whepSession.RegisterWhepHandlers(peerConnection)

	session.WhepSessionsLock.Lock()
	session.WhepSessions[whepSessionId] = whepSession
	session.WhepSessionsLock.Unlock()

	go session.handleWhepConnection(host, whepSession)
	go session.handleWhepChannels(whepSession)
	go session.handleWhepVideoRtcpSender(videoRtcpSender)

	// TODO: Implement
	// go session.handleWhepLayerChange(host, whepSession)

	return nil
}

// Add host
func (session *Session) AddHost(peerConnection *webrtc.PeerConnection) (err error) {
	log.Println("Session.AddHost")

	for {
		host := session.Host.Load()
		if host == nil {
			break
		}

		if host.PeerConnection.ConnectionState() != webrtc.PeerConnectionStateClosed || session.ActiveContext.Err() == nil {
			return fmt.Errorf("session already has a host")
		}

		if session.Host.CompareAndSwap(host, nil) {
			break
		}
	}

	activeContext, activeContextCancel := context.WithCancel(context.Background())

	host := &whip.WhipSession{
		Id:                          uuid.New().String(),
		AudioTracks:                 make(map[string]*whip.AudioTrack),
		VideoTracks:                 make(map[string]*whip.VideoTrack),
		PacketLossIndicationChannel: make(chan bool, 50),
		OnTrackChangeChannel:        make(chan struct{}, 50),
		EventsChannel:               make(chan any, 50),

		ActiveContext:       activeContext,
		ActiveContextCancel: activeContextCancel,
	}

	host.AddPeerConnection(peerConnection, session.StreamKey)
	if !session.Host.CompareAndSwap(nil, host) {
		host.ActiveContextCancel()
		host.RemovePeerConnection()
		host.RemoveTracks()
		return fmt.Errorf("session already has a host")
	}

	go session.hostStatusLoop()

	return nil
}

func (session *Session) RemoveHost() {

	host := session.Host.Swap(nil)
	if host == nil {
		log.Println("Session.RemoveHost", session.StreamKey, "- No host to remove")
		return
	}

	log.Println("Session.RemoveHost", session.StreamKey)

	host.ActiveContextCancel()
	host.RemovePeerConnection()
	host.RemoveTracks()
}

// Remove Whep session from Whip session
// In case the Whip session does not have a host, and no more whep sessions, it will
// be remove from the manager.
func (session *Session) removeWhep(whepSessionId string) {
	log.Println("Session.RemoveWhepSession:", session.StreamKey, " - ", whepSessionId)

	session.WhepSessionsLock.Lock()
	session.WhepSessions[whepSessionId].Close()
	delete(session.WhepSessions, whepSessionId)
	session.WhepSessionsLock.Unlock()

	if session.isEmpty() {
		session.close()
	}
}

// Remove all Hosts and clients before closing down session
func (session *Session) close() {

	session.WhepSessionsLock.Lock()
	for _, whep := range session.WhepSessions {
		whep.Close()
	}
	session.WhepSessions = make(map[string]*whep.WhepSession)
	session.WhepSessionsLock.Unlock()

	session.RemoveHost()

	session.ActiveContextCancel()
}

func (session *Session) Close() {
	log.Println("Session.Close", session.StreamKey)
	session.close()
}

// Returns true is no WHIP tracks are present, and no WHEP sessions are waiting for incoming streams
func (session *Session) isEmpty() bool {
	if session.hasWhepSessions() {
		log.Println("Session.IsEmpty.HasWhepSessions (false):", session.StreamKey)
		return false
	}

	if session.isStreaming() {
		log.Println("Session.IsEmpty.IsActive (false):", session.StreamKey)
		return false
	}

	log.Println("Session.IsEmpty (true):", session.StreamKey)
	return true
}

// Returns true if any tracks are available for the session
func (session *Session) isStreaming() bool {

	host := session.Host.Load()
	if host == nil {
		return false
	}

	host.TracksLock.RLock()

	if len(host.AudioTracks) != 0 {
		log.Println("Session.IsActive.AudioTracks", len(host.AudioTracks))
		host.TracksLock.RUnlock()
		return true
	}
	if len(host.VideoTracks) != 0 {
		log.Println("Session.IsActive.VideoTracks", len(host.VideoTracks))
		host.TracksLock.RUnlock()
		return true
	}

	host.TracksLock.RUnlock()
	return false
}

func (session *Session) hasWhepSessions() bool {
	session.WhepSessionsLock.RLock()
	log.Println("Session.HasWhepSessions:", len(session.WhepSessions))

	if len(session.WhepSessions) == 0 {
		session.WhepSessionsLock.RUnlock()
		return false
	}

	session.WhepSessionsLock.RUnlock()
	return true
}

// Get the status of the current session
func (session *Session) GetStreamStatus() (status WhipSessionStatus) {
	session.WhepSessionsLock.RLock()
	whepSessionsCount := len(session.WhepSessions)
	session.WhepSessionsLock.RUnlock()

	session.StatusLock.RLock()

	status = WhipSessionStatus{
		StreamKey:   session.StreamKey,
		MOTD:        session.MOTD,
		ViewerCount: whepSessionsCount,
		IsOnline:    session.HasHost.Load(),
		StreamStart: session.StreamStart,
	}

	session.StatusLock.RUnlock()

	return
}
