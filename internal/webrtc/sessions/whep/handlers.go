package whep

import (
	"log/slog"

	"github.com/pion/webrtc/v4"
)

func (w *WHEPSession) RegisterWHEPHandlers(peerConnection *webrtc.PeerConnection) {
	slog.Info("WHEPSession.RegisterHandlers")

	peerConnection.OnICEConnectionStateChange(onWHEPICEConnectionStateChangeHandler(w))
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
