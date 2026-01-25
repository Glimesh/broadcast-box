package whep

import (
	"log"

	"github.com/pion/webrtc/v4"
)

func (whep *WhepSession) RegisterWhepHandlers(peerConnection *webrtc.PeerConnection) {
	log.Println("WhepSession.RegisterHandlers")

	peerConnection.OnICEConnectionStateChange(onWhepICEConnectionStateChangeHandler(whep))
}

func onWhepICEConnectionStateChangeHandler(whep *WhepSession) func(webrtc.ICEConnectionState) {
	return func(state webrtc.ICEConnectionState) {
		log.Println("WhepSession.OnICEConnectionStateChange:", state)
		switch state {
		case
			webrtc.ICEConnectionStateConnected:
			select {
			case whep.ConnectionChannel <- true:
			default:
			}
		case
			webrtc.ICEConnectionStateFailed,
			webrtc.ICEConnectionStateClosed:
			whep.Close()
		default:
			log.Println("WhepSession.OnICEConnectionStateChange.Default", state)
		}
	}
}
