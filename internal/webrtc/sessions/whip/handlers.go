package whip

import (
	"log"
	"strings"

	"github.com/pion/webrtc/v4"
)

func (whip *WhipSession) RegisterWhipHandlers(peerConnection *webrtc.PeerConnection, streamKey string) {
	log.Println("WhipSession.RegisterHandlers")

	// PeerConnection OnTrack handler
	whip.PeerConnection.OnTrack(whip.onTrackHandler(peerConnection, streamKey))

	// PeerConnection OnICEConnectionStateChange handler
	whip.PeerConnection.OnICEConnectionStateChange(whip.onICEConnectionStateChangeHandler())

	// PeerConnection OnConnectionStateChange
	whip.PeerConnection.OnConnectionStateChange(whip.onConnectionStateChange())
}

func (whip *WhipSession) onICEConnectionStateChangeHandler() func(webrtc.ICEConnectionState) {
	return func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateFailed || state == webrtc.ICEConnectionStateClosed {
			log.Println("WhipSession.PeerConnection.OnICEConnectionStateChange", whip.Id)
			whip.ActiveContextCancel()
		}
	}
}

func (whip *WhipSession) onTrackHandler(peerConnection *webrtc.PeerConnection, streamKey string) func(*webrtc.TrackRemote, *webrtc.RTPReceiver) {
	return func(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
		log.Println("WhipSession.PeerConnection.OnTrackHandler", whip.Id)
		whip.OnTrackChangeChannel <- struct{}{}

		if strings.HasPrefix(remoteTrack.Codec().MimeType, "audio") {
			// Handle audio stream
			whip.AudioWriter(remoteTrack, streamKey, peerConnection)
		} else {
			// Handle video stream
			whip.VideoWriter(remoteTrack, streamKey, peerConnection)
		}

		// Fires when track has stopped
		whip.OnTrackChangeChannel <- struct{}{}

		log.Println("WhipSession.OnTrackHandler.TrackStopped", remoteTrack.RID())
	}
}

func (whip *WhipSession) onConnectionStateChange() func(webrtc.PeerConnectionState) {
	return func(state webrtc.PeerConnectionState) {
		log.Println("WhipSession.PeerConnection.OnConnectionStateChange", state)

		switch state {
		case webrtc.PeerConnectionStateClosed:
		case webrtc.PeerConnectionStateFailed:
			log.Println("WhipSession.PeerConnection.OnConnectionStateChange: Host removed", whip.Id)
			whip.ActiveContextCancel()

		case webrtc.PeerConnectionStateConnected:
			log.Println("WhipSession.PeerConnection.OnConnectionStateChange: Host connected", whip.Id)

		}
	}
}
