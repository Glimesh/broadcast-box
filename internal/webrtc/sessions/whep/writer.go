package whep

import (
	"errors"
	"io"
	"log"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
)

// Sends provided audio packet to the Whep session
func (whepSession *WhepSession) SendAudioPacket(packet codecs.TrackPacket) {
	if whepSession.IsSessionClosed.Load() || whepSession.AudioTrack == nil {
		return
	}

	whepSession.AudioLock.Lock()
	whepSession.AudioPacketsWritten += 1
	whepSession.AudioTimestamp = uint32(int64(whepSession.AudioTimestamp) + packet.TimeDiff)
	whepSession.AudioLock.Unlock()

	if err := whepSession.AudioTrack.WriteRTP(packet.Packet, packet.Codec); err != nil {
		if errors.Is(err, io.ErrClosedPipe) {
			log.Println("WhepSession.SendAudioPacket.ConnectionDropped")
			whepSession.Close()
		} else {
			log.Println("WhepSession.SendAudioPacket.Error", err)
		}
	}
}

// Sends provided video packet to the Whep session
func (whepSession *WhepSession) SendVideoPacket(packet codecs.TrackPacket) {
	if whepSession.IsSessionClosed.Load() {
		return
	}

	if whepSession.IsWaitingForKeyframe.Load() {
		if !packet.IsKeyframe {
			return
		}

		whepSession.IsWaitingForKeyframe.Store(false)
	}

	whepSession.VideoLock.Lock()
	whepSession.VideoBytesWritten += len(packet.Packet.Payload)
	whepSession.VideoPacketsWritten += 1
	whepSession.VideoSequenceNumber = uint16(whepSession.VideoSequenceNumber) + uint16(packet.SequenceDiff)
	whepSession.VideoTimestamp = uint32(int64(whepSession.VideoTimestamp) + packet.TimeDiff)
	whepSession.VideoLock.Unlock()

	packet.Packet.SequenceNumber = whepSession.VideoSequenceNumber
	packet.Packet.Timestamp = whepSession.VideoTimestamp

	if err := whepSession.VideoTrack.WriteRTP(packet.Packet, packet.Codec); err != nil {
		whepSession.VideoPacketsDropped.Add(1)

		if errors.Is(err, io.ErrClosedPipe) {
			log.Println("WhepSession.SendVideoPacket.ConnectionDropped")
			whepSession.Close()
		} else {
			log.Println("WhepSession.SendVideoPacket.Error", err)
		}
	}
}
