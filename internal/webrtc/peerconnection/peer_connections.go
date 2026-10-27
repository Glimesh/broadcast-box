package peerconnection

import (
	"log"

	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
	"github.com/pion/webrtc/v4"
)

func CreateWHEPPeerConnection() (*webrtc.PeerConnection, error) {
	return manager.APIWHEP.NewPeerConnection(getPeerConnectionConfig())
}

func CreateWHIPPeerConnection(offer string) (*webrtc.PeerConnection, error) {
	log.Println("CreateWHIPPeerConnection.CreateWHIPPeerConnection")

	peerConnection, err := manager.APIWHIP.NewPeerConnection(getPeerConnectionConfig())
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
	log.Println("PeerConnection.CreateWHIPPeerConnection.GatheringCompleteResult")

	return peerConnection, nil
}
