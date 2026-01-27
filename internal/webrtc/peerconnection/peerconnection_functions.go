package peerconnection

import (
	"log"

	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
	"github.com/pion/webrtc/v4"
)

type CreateWhipPeerConnectionResult struct {
	PeerConnection *webrtc.PeerConnection
	Error          error
}

func CreateWhepPeerConnection() (*webrtc.PeerConnection, error) {
	return manager.ApiWhep.NewPeerConnection(GetPeerConnectionConfig())
}

func CreateWhipPeerConnection(offer string) (*webrtc.PeerConnection, error) {
	log.Println("CreateWhipPeerConnection.CreateWhipPeerConnection")

	peerConnection, err := manager.ApiWhip.NewPeerConnection(GetPeerConnectionConfig())
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
