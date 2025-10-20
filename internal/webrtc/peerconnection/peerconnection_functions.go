package peerconnection

import (
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
	"github.com/pion/webrtc/v4"
	"log"
)

type CreateWhipPeerConnectionResult struct {
	PeerConnection *webrtc.PeerConnection
	Error          error
}

func CreateWhepPeerConnection() (*webrtc.PeerConnection, error) {
	return session.ApiWhep.NewPeerConnection(GetPeerConnectionConfig())
}

func CreateWhipPeerConnection(offer string) (*webrtc.PeerConnection, error) {
	log.Println("CreateWhipPeerConnection.CreateWhipPeerConnection")

	peerConnection, err := session.ApiWhip.NewPeerConnection(GetPeerConnectionConfig())
	if err != nil {
		return nil, err
	}

	// Setup PeerConnection RemoteDescription
	sessionDescription := webrtc.SessionDescription{
		SDP:  string(offer),
		Type: webrtc.SDPTypeOffer,
	}

	if err := peerConnection.SetRemoteDescription(sessionDescription); err != nil {
		return nil, err
	}

	gatheringCompleteResult := webrtc.GatheringCompletePromise(peerConnection)

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err := peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	// Await gathering trickle
	<-gatheringCompleteResult
	log.Println("PeerConnection.CreateWhipPeerConnection.GatheringCompleteResult")

	return peerConnection, nil
}

func disconnected(isWhip bool, streamKey string, sessionId string) {
	whipSession, ok := session.SessionManager.GetWhipStream(streamKey)

	if isWhip {
		log.Println("WhipSession.Disconnected:", streamKey, "found was", ok)
	} else {
		log.Println("WhepSession.Disconnected:", streamKey, "found was", ok)
	}

	if !ok {
		return
	}

	// Remove active tracks if it is a WHIP session
	if isWhip {
		log.Println("WhipSession.Disconnected: Removing tracks", sessionId)
		whipSession.RemoveTracks()
	} else {
		log.Println("WhepSession.Disconnected: Removing session", sessionId)
		//TODO: Find a way to use the SSE connection to be considdered the tether to an open connection
		// whipSession.RemoveWhepSession(sessionId)
	}

	// Do not conclude stream if whep sessions are still listening, or the host is still active
	if whipSession.HasWhepSessions() {
		return
	}

	// Remove Whip session from manager if its empty
	log.Println("WhipSession.RemoveWhipSession: No Whep session, closing down")
	whipSession.ActiveContextCancel()
}
