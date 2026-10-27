package whip

import (
	"log"
	"strings"

	"github.com/glimesh/broadcast-box/internal/webrtc/chatdc"
	"github.com/pion/webrtc/v4"
)

func (w *WHIPSession) registerWHIPHandlers(peerConnection *webrtc.PeerConnection, streamKey string) {
	log.Println("WHIPSession.RegisterHandlers")

	// PeerConnection OnTrack handler
	w.PeerConnection.OnTrack(w.onTrackHandler(peerConnection, streamKey))

	// PeerConnection OnICEConnectionStateChange handler
	w.PeerConnection.OnICEConnectionStateChange(w.onICEConnectionStateChangeHandler())

	// PeerConnection OnConnectionStateChange
	w.PeerConnection.OnConnectionStateChange(w.onConnectionStateChange())

	// PeerConnection DataChannel chat handler
	w.PeerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		handler := chatdc.NewHandler(w.ChatManager)
		handler.Bind(streamKey, w.ID, dataChannel)
	})
}

func (w *WHIPSession) onICEConnectionStateChangeHandler() func(webrtc.ICEConnectionState) {
	return func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateFailed || state == webrtc.ICEConnectionStateClosed {
			log.Println("WHIPSession.PeerConnection.OnICEConnectionStateChange", w.ID)
			w.notifyClosed()
		}
	}
}

func (w *WHIPSession) onTrackHandler(peerConnection *webrtc.PeerConnection, streamKey string) func(*webrtc.TrackRemote, *webrtc.RTPReceiver) {
	return func(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
		log.Println("WHIPSession.PeerConnection.OnTrackHandler", w.ID)

		if strings.HasPrefix(remoteTrack.Codec().MimeType, "audio") {
			// Handle audio stream
			w.audioWriter(remoteTrack, streamKey)
		} else {
			// Handle video stream
			w.videoWriter(remoteTrack, streamKey, peerConnection)
		}

		log.Println("WHIPSession.OnTrackHandler.TrackStopped", remoteTrack.RID())
	}
}

func (w *WHIPSession) onConnectionStateChange() func(webrtc.PeerConnectionState) {
	return func(state webrtc.PeerConnectionState) {
		log.Println("WHIPSession.PeerConnection.OnConnectionStateChange", state)

		switch state {
		case webrtc.PeerConnectionStateClosed:
			w.notifyClosed()
		case webrtc.PeerConnectionStateFailed:
			log.Println("WHIPSession.PeerConnection.OnConnectionStateChange: Host removed", w.ID)
			w.notifyClosed()

		case webrtc.PeerConnectionStateConnected:
			log.Println("WHIPSession.PeerConnection.OnConnectionStateChange: Host connected", w.ID)

		}
	}
}
