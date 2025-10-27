package whep

import (
	"errors"
	"io"
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
		AudioChannel:         make(chan TrackPacket, 100),
		VideoChannel:         make(chan TrackPacket, 100),
		WhipEventsChannel:    make(chan any, 100),
		SseEventsChannel:     make(chan any, 100),
		SessionClosedChannel: make(chan struct{}, 1),
		PeerConnection:       peerConnection,
	}

	whepSession.AudioLayerCurrent.Store(audioLayer)
	whepSession.VideoLayerCurrent.Store(videoLayer)
	whepSession.IsWaitingForKeyframe.Store(false)
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

	// Handle WHEP Incoming stream
	go func() {
		var lastPacketSequence uint16 = 0
		for {
			select {
			case packet := <-whepSession.VideoChannel:

				if lastPacketSequence < packet.Packet.SequenceNumber+uint16(packet.SequenceDiff) {
					lastPacketSequence = packet.Packet.SequenceNumber

					// Convert to WhepSession Function
					whepSession.VideoLock.Lock()
					whepSession.VideoPacketsWritten += 1
					whepSession.VideoSequenceNumber = uint16(whepSession.VideoSequenceNumber) + uint16(packet.SequenceDiff)
					whepSession.VideoTimestamp = uint32(int64(whepSession.VideoTimestamp) + packet.TimeDiff)

					packet.Packet.SequenceNumber = whepSession.VideoSequenceNumber
					packet.Packet.Timestamp = whepSession.VideoTimestamp
					whepSession.VideoLock.Unlock()

					if whepSession.IsWaitingForKeyframe.Load() {
						if !packet.IsKeyframe {
							log.Println("WhepSession.SendVideoPacket: Waiting for keyframe")
							return
						}
					}

					whepSession.IsWaitingForKeyframe.Store(false)

					if err := whepSession.VideoTrack.WriteRTP(packet.Packet, packet.Codec); err != nil {
						if errors.Is(err, io.ErrClosedPipe) {
							log.Println("WhepSession.SendVideoPacket.ConnectionDropped")
							whepSession.Close()
						} else {
							log.Println("WhepSession.SendVideoPacket.Error", err)
						}
					}
				}
			case packet := <-whepSession.AudioChannel:
				whepSession.SendAudioPacket(packet)
			case <-whepSession.SessionClosedChannel:
				return
			}
		}
	}()

	return whepSession
}
