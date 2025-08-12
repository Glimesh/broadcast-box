package peerconnection

import (
	"log"
	"strings"

	"github.com/glimesh/broadcast-box/internal/webrtc/session"
	"github.com/glimesh/broadcast-box/internal/webrtc/session/track"
	"github.com/pion/webrtc/v4"
)

func RegisterHandlers(peerConnection *webrtc.PeerConnection, stream *session.WhipSession, isWhipSession bool, sessionId string) {
	// PeerConnection Handlers
	peerConnection.OnTrack(onTrackHandler(stream, peerConnection))
	peerConnection.OnICEConnectionStateChange(
		onICEConnectionStateChangeHandler(
			peerConnection,
			isWhipSession,
			stream.StreamKey,
			sessionId))
}

func onICEConnectionStateChangeHandler(
	peerConnection *webrtc.PeerConnection,
	isWhip bool,
	streamKey string,
	sessionId string,
) func(webrtc.ICEConnectionState) {

	return func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateFailed || state == webrtc.ICEConnectionStateClosed {
			if err := peerConnection.Close(); err != nil {
				log.Println("PeerConnection.OnICEConnectionStateChange.Error", err)
			}

			if strings.EqualFold("DEBUG_PEERCONNECTION_ENABLED", "true") {
				if isWhip {
					log.Println("WHIP: Disconnected", streamKey, sessionId)
				} else {
					log.Println("WHEP: Disconnected", streamKey, sessionId)
				}
			}

			disconnected(isWhip, streamKey, sessionId)
		}
	}
}

func onTrackHandler(stream *session.WhipSession, peerConnection *webrtc.PeerConnection) func(*webrtc.TrackRemote, *webrtc.RTPReceiver) {
	return func(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
		if strings.HasPrefix(remoteTrack.Codec().MimeType, "audio") {
			// Handle audio stream
			track.AudioWriter(remoteTrack, stream, peerConnection)
		} else {
			// Handle video stream
			track.VideoWriter(remoteTrack, stream, peerConnection)
		}
	}
}
