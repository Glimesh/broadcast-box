package peerconnection

import (
	"log"
	"strings"

	"github.com/glimesh/broadcast-box/internal/webrtc/session/whip"
	"github.com/pion/webrtc/v4"
)

func RegisterWhipHandlers(whipSession *whip.WhipSession, peerConnection *webrtc.PeerConnection, sessionId string) {
	log.Println("WhipSession.RegisterHandlers")

	// PeerConnection OnTrack handler
	whipSession.PeerConnection.OnTrack(onWhipTrackHandler(whipSession, peerConnection))

	// PeerConnection OnICEConnectionStateChange handler
	whipSession.PeerConnection.OnICEConnectionStateChange(
		onWhipICEConnectionStateChangeHandler(
			whipSession.StreamKey,
			sessionId))

	// PeerConnection OnConnectionStateChange
	whipSession.PeerConnection.OnConnectionStateChange(onConnectionStateChange(whipSession))
}

func onWhipICEConnectionStateChangeHandler(
	streamKey string,
	sessionId string,
) func(webrtc.ICEConnectionState) {

	return func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateFailed || state == webrtc.ICEConnectionStateClosed {

			if strings.EqualFold("DEBUG_PEERCONNECTION_ENABLED", "true") {
				log.Println("WhepSession: Disconnected", streamKey, sessionId)
			}

			disconnected(true, streamKey, sessionId)
		}
	}
}

func onWhipTrackHandler(whipSession *whip.WhipSession, peerConnection *webrtc.PeerConnection) func(*webrtc.TrackRemote, *webrtc.RTPReceiver) {
	return func(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
		log.Println("WhipSession.PeerConnection.OnTrackHandler", whipSession.StreamKey)
		whipSession.OnTrackChangeChannel <- struct{}{}
		if strings.HasPrefix(remoteTrack.Codec().MimeType, "audio") {
			// Handle audio stream
			whipSession.AudioWriter(remoteTrack, peerConnection)
		} else {
			// Handle video stream
			whipSession.VideoWriter(remoteTrack, peerConnection)
		}

		// Fires when track has stopped
		whipSession.OnTrackChangeChannel <- struct{}{}
		peerConnection.Close()

		log.Println("WhipSession.onWhipTrackHandler.TrackStopped", remoteTrack.RID())
	}
}

func onConnectionStateChange(whipSession *whip.WhipSession) func(webrtc.PeerConnectionState) {
	return func(state webrtc.PeerConnectionState) {
		log.Println("WhipSession.PeerConnection.OnConnectionStateChange", state)

		if state == webrtc.PeerConnectionStateDisconnected || state == webrtc.PeerConnectionStateClosed || state == webrtc.PeerConnectionStateFailed {
			log.Println("WhipSession.PeerConnection.OnConnectionStateChange: Host removed")
			whipSession.ActiveContextCancel()
			whipSession.HasHost.Store(false)
		}
	}
}
