package whep

import (
	"context"
	"log"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/webrtc/v4"
)

func CreateNewWhep(whepSessionId string, audioTrack *codecs.TrackMultiCodec, audioLayer string, videoTrack *codecs.TrackMultiCodec, videoLayer string, peerConnection *webrtc.PeerConnection) (whepSession *WhepSession) {
	log.Println("WhepSession.CreateNewWhep", whepSessionId)

	activeContext, activeContextCancel := context.WithCancel(context.Background())
	whepSession = &WhepSession{
		SessionId:           whepSessionId,
		AudioTrack:          audioTrack,
		VideoTrack:          videoTrack,
		AudioTimestamp:      5000,
		VideoTimestamp:      5000,
		AudioChannel:        make(chan codecs.TrackPacket, 2500),
		VideoChannel:        make(chan codecs.TrackPacket, 2500),
		WhipEventsChannel:   make(chan any, 100),
		SseEventsChannel:    make(chan any, 100),
		PeerConnection:      peerConnection,
		ActiveContext:       activeContext,
		ActiveContextCancel: activeContextCancel,
	}

	log.Println("WhepSession.CreateNewWhep.AudioLayer", audioLayer)
	log.Println("WhepSession.CreateNewWhep.VideoLayer", videoLayer)
	whepSession.AudioLayerCurrent.Store(audioLayer)
	whepSession.VideoLayerCurrent.Store(videoLayer)
	whepSession.IsWaitingForKeyframe.Store(true)
	whepSession.IsSessionClosed.Store(false)

	// Handle WHEP Streams and Events
	go whepSession.handleEvents()
	go whepSession.handleStream()

	return whepSession
}

// Handle events for SSE to the WHEP sessions
func (whepSession *WhepSession) handleEvents() {
	for {
		select {
		case <-whepSession.ActiveContext.Done():
			log.Println("WhepSession.HandleEventsLoop.Close")
			return
		case msg, ok := <-whepSession.WhipEventsChannel:
			if !ok {
				log.Println("WhepSession.Event.Whip: Channel closed - exiting")
			} else {
				log.Println("WhepSession.Event.Whip:", msg)
			}
		}
	}
}

// Handles incoming stream packets
func (whepSession *WhepSession) handleStream() {
	for {
		select {
		case <-whepSession.ActiveContext.Done():
			log.Println("WhepSession.HandleStreamLoop.Close")
			return

		case packet := <-whepSession.VideoChannel:
			if packet.Layer == whepSession.VideoLayerCurrent.Load() {
				whepSession.SendVideoPacket(packet)
			}

		case packet := <-whepSession.AudioChannel:
			if packet.Layer == whepSession.AudioLayerCurrent.Load() {
				whepSession.SendAudioPacket(packet)
			}
		}
	}
}
