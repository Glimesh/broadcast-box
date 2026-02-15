package whip

import (
	"log"

	"github.com/pion/webrtc/v4"
)

func (whip *WhipSession) AddPeerConnection(peerConnection *webrtc.PeerConnection, streamKey string) {
	log.Println("WhipSession.AddPeerConnection")

	whip.PeerConnectionLock.Lock()
	existingPeerConnection := whip.PeerConnection
	whip.PeerConnection = peerConnection
	whip.PeerConnectionLock.Unlock()

	if existingPeerConnection != nil && existingPeerConnection != peerConnection {
		log.Println("WhipSession.AddPeerConnection: Replacing existing peerconnection")
		if err := existingPeerConnection.GracefulClose(); err != nil {
			log.Println("WhipSession.AddPeerConnection.Close.Error", err)
		}
	}

	whip.RegisterWhipHandlers(peerConnection, streamKey)
}

func (whip *WhipSession) RemovePeerConnection() {
	log.Println("WhipSession.RemovePeerConnection", whip.Id)

	whip.PeerConnectionLock.Lock()
	peerConnection := whip.PeerConnection
	whip.PeerConnection = nil
	whip.PeerConnectionLock.Unlock()

	if peerConnection == nil {
		return
	}

	if err := peerConnection.Close(); err != nil {
		log.Println("WhipSession.RemovePeerConnection.Error", err)
	}

	log.Println("WhipSession.RemovePeerConnection.Completed", whip.Id)
}
