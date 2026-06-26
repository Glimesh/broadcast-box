package whep

import (
	"log/slog"

	"github.com/glimesh/broadcast-box/internal/webrtc/chatdc"
	"github.com/glimesh/broadcast-box/internal/webrtc/datadc"
	"github.com/pion/webrtc/v4"
)

func (w *WHEPSession) RegisterWHEPHandlers(peerConnection *webrtc.PeerConnection, peers datadc.PeerStore) {
	slog.Info("WHEPSession.RegisterHandlers")

	peerConnection.OnICEConnectionStateChange(onWHEPICEConnectionStateChangeHandler(w))

	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		chatHandler := chatdc.NewHandler(w.ChatManager)
		chatHandler.Bind(w.StreamKey, w.SessionID, dataChannel)
		datadc.Bind(w.StreamKey, peers, w.SessionID, dataChannel)
	})
}

func onWHEPICEConnectionStateChangeHandler(w *WHEPSession) func(webrtc.ICEConnectionState) {
	return func(state webrtc.ICEConnectionState) {
		slog.Info("WHEPSession.OnICEConnectionStateChange", "state", state)
		switch state {
		case
			webrtc.ICEConnectionStateConnected:
			w.SendPLI()
		case
			webrtc.ICEConnectionStateFailed,
			webrtc.ICEConnectionStateClosed:
			w.Close()
		default:
			slog.Info("WHEPSession.OnICEConnectionStateChange.Default", "state", state)
		}
	}
}
