package whip

import (
	"log"

	"github.com/pion/webrtc/v4"
)

func (whipSession *WhipSession) AddPeerConnection(peerConnection *webrtc.PeerConnection) {
	log.Println("WhipSession.AddPeerConnection")

	isActive := whipSession.IsActive()
	log.Println("WhipSession.AddPeerConnection.IsActive:", isActive)

	whipSession.PeerConnectionLock.Lock()
	if !isActive || whipSession.PeerConnection == nil {
		whipSession.PeerConnection = peerConnection
	} else {
		log.Println("WhipSession.AddPeerConnection: A PeerConnection already exists")
		whipSession.PeerConnection = nil
		whipSession.PeerConnection = peerConnection
	}
	whipSession.PeerConnectionLock.Unlock()

	log.Println("WhipSession.AddPeerConnection.Complete")
}

func (whipSession *WhipSession) RemovePeerConnection() {
	log.Println("WhipSession.RemovePeerConnection", whipSession.StreamKey)

	err := whipSession.PeerConnection.Close()
	if err != nil {
		log.Println("WhipSession.RemovePeerConnection.Error", err)
	}

	whipSession.PeerConnectionLock.Lock()
	whipSession.PeerConnection = nil
	whipSession.PeerConnectionLock.Unlock()

	log.Println("WhipSession.RemovePeerConnection.Completed", whipSession.StreamKey)
}
