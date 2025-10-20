package peerconnection

import (
	"log"
	"strings"

	"github.com/glimesh/broadcast-box/internal/webrtc/session/whip"
	"github.com/pion/webrtc/v4"
)

func RegisterWhepHandlers(whipSession *whip.WhipSession, peerConnection *webrtc.PeerConnection, sessionId string) {
	log.Println("PeerConnection.RegisterHandlers")

	whipSession.PeerConnectionLock.RLock()
	whipHasPeerConnection := whipSession.PeerConnection != nil
	whipSession.PeerConnectionLock.RUnlock()

	if !whipHasPeerConnection {
		return
	}

	// PeerConnection OnICEConnectionStateChange handler
	whipSession.PeerConnection.OnICEConnectionStateChange(
		onWhepICEConnectionStateChangeHandler(
			whipSession.StreamKey,
			sessionId))
}

func onWhepICEConnectionStateChangeHandler(
	streamKey string,
	sessionId string,
) func(webrtc.ICEConnectionState) {

	return func(state webrtc.ICEConnectionState) {
		log.Println("WhepSession.PeerConnection.State", state, streamKey)
		if state == webrtc.ICEConnectionStateFailed || state == webrtc.ICEConnectionStateClosed {

			if strings.EqualFold("DEBUG_PEERCONNECTION_ENABLED", "true") {
				log.Println("WhepSession: Disconnected", streamKey, sessionId)
			}

			disconnected(false, streamKey, sessionId)
		}
	}
}
