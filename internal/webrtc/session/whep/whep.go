package whep

import (
	"log"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/webrtc/v4"
)

func CreateNewWhep(whepSessionId string, audioTrack *codecs.TrackMultiCodec, audioLayer string, videoTrack *codecs.TrackMultiCodec, videoLayer string, peerConnection *webrtc.PeerConnection) (whepSession *WhepSession) {
	log.Println("WhepSession.CreateNewWhep", whepSessionId)

	whepSession = &WhepSession{
		SessionId:            whepSessionId,
		AudioTrack:           audioTrack,
		VideoTrack:           videoTrack,
		AudioTimestamp:       5000,
		VideoTimestamp:       5000,
		AudioChannel:         make(chan codecs.TrackPacket, 1000),
		VideoChannel:         make(chan codecs.TrackPacket, 1000),
		WhipEventsChannel:    make(chan any, 100),
		SseEventsChannel:     make(chan any, 100),
		SessionClosedChannel: make(chan struct{}, 1),
		PeerConnection:       peerConnection,
	}

	log.Println("WhepSession.CreateNewWhep.AudioLayer", audioLayer)
	log.Println("WhepSession.CreateNewWhep.VideoLayer", videoLayer)
	whepSession.AudioLayerCurrent.Store(audioLayer)
	whepSession.VideoLayerCurrent.Store(videoLayer)
	whepSession.IsWaitingForKeyframe.Store(true)
	whepSession.IsSessionClosed.Store(false)

	// Handle WHEP Events
	go func() {
		for {
			select {
			case msg, ok := <-whepSession.WhipEventsChannel:
				if !ok {
					log.Println("WhepSession.Event.Whip: Channel closed - exiting")
				} else {
					log.Println("WhepSession.Event.Whip:", msg)
				}
			case <-whepSession.SessionClosedChannel:
				return
			}
		}
	}()

	// Handle WHEP Incoming streams
	go func() {
		// TODO: Include a timer to check for change in current audio/video layer
		currentAudioLayer := whepSession.AudioLayerCurrent.Load()
		currentVideoLayer := whepSession.VideoLayerCurrent.Load()

		for {
			select {
			case <-whepSession.SessionClosedChannel:
				return

			case packet := <-whepSession.VideoChannel:
				if packet.Layer == currentVideoLayer {
					whepSession.SendVideoPacket(packet)
				}

			case packet := <-whepSession.AudioChannel:
				if packet.Layer == currentAudioLayer {
					whepSession.SendAudioPacket(packet)
				}
			}
		}
	}()

	return whepSession
}
