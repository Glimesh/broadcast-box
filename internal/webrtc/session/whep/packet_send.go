package whep

import (
	"errors"
	"io"
	"log"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
)

// Sends provided audio packet to the Whep session
func (whepSession *WhepSession) SendAudioPacket(packet codecs.TrackPacket) {
	whepSession.AudioLock.RLock()
	audioTrack := whepSession.AudioTrack
	whepSession.AudioLock.RUnlock()

	if whepSession.IsSessionClosed.Load() || audioTrack == nil {
		return
	}

	whepSession.AudioLock.Lock()
	whepSession.AudioPacketsWritten += 1
	whepSession.AudioSequenceNumber = uint16(whepSession.AudioSequenceNumber) + uint16(packet.SequenceDiff)
	whepSession.AudioTimestamp = uint32(int64(whepSession.AudioTimestamp) + packet.TimeDiff)

	packet.Packet.SequenceNumber = whepSession.AudioSequenceNumber
	packet.Packet.Timestamp = whepSession.AudioTimestamp
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
		log.Println("WhepSession.SendVideoPacket.SessionClosed")
		return
	}

	if whepSession.IsWaitingForKeyframe.Load() {
		if !packet.IsKeyframe {
			return
		}

		whepSession.IsWaitingForKeyframe.Store(false)
	}

	whepSession.VideoLock.Lock()
	whepSession.VideoPacketsWritten += 1
	whepSession.VideoSequenceNumber = uint16(whepSession.VideoSequenceNumber) + uint16(packet.SequenceDiff)
	whepSession.VideoTimestamp = uint32(int64(whepSession.VideoTimestamp) + packet.TimeDiff)

	packet.Packet.SequenceNumber = whepSession.VideoSequenceNumber
	packet.Packet.Timestamp = whepSession.VideoTimestamp
	whepSession.VideoLock.Unlock()

	if err := whepSession.VideoTrack.WriteRTP(packet.Packet, packet.Codec); err != nil {
		if errors.Is(err, io.ErrClosedPipe) {
			log.Println("WhepSession.SendVideoPacket.ConnectionDropped")
			whepSession.Close()
		} else {
			log.Println("WhepSession.SendVideoPacket.Error", err)
		}
	}
}
