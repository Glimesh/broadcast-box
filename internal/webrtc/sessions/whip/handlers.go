package whip

import (
	"log/slog"
	"strings"

	"github.com/glimesh/broadcast-box/internal/webrtc/chatdc"
	"github.com/pion/webrtc/v4"
)

func (w *WHIPSession) registerWHIPHandlers(peerConnection *webrtc.PeerConnection, streamKey string) {
	slog.Info("WHIPSession.RegisterHandlers")

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
			slog.Info("WHIPSession.PeerConnection.OnICEConnectionStateChange", "id", w.ID)
			w.notifyClosed()
		}
	}
}

func (w *WHIPSession) onTrackHandler(peerConnection *webrtc.PeerConnection, streamKey string) func(*webrtc.TrackRemote, *webrtc.RTPReceiver) {
	return func(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
		slog.Info("WHIPSession.PeerConnection.OnTrackHandler", "id", w.ID)

		if strings.HasPrefix(remoteTrack.Codec().MimeType, "audio") {
			// Handle audio stream
			w.audioWriter(remoteTrack, streamKey)
		} else {
			// Handle video stream
			w.videoWriter(remoteTrack, streamKey, peerConnection)
		}

		slog.Info("WHIPSession.OnTrackHandler.TrackStopped", "rid", remoteTrack.RID())
	}
}

func (w *WHIPSession) onConnectionStateChange() func(webrtc.PeerConnectionState) {
	return func(state webrtc.PeerConnectionState) {
		slog.Info("WHIPSession.PeerConnection.OnConnectionStateChange", "state", state)

		switch state {
		case webrtc.PeerConnectionStateClosed:
			w.notifyClosed()
		case webrtc.PeerConnectionStateFailed:
			slog.Info("WHIPSession.PeerConnection.OnConnectionStateChange: Host removed", "id", w.ID)
			w.notifyClosed()

		case webrtc.PeerConnectionStateConnected:
			slog.Info("WHIPSession.PeerConnection.OnConnectionStateChange: Host connected", "id", w.ID)

		}
	}
}
