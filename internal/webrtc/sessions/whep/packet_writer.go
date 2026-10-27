package whep

import (
	"errors"
	"io"
	"log"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
)

// Sends provided audio packet to the WHEP session
func (w *WHEPSession) SendAudioPacket(packet codecs.TrackPacket) {
	if w.IsSessionClosed.Load() {
		return
	}

	w.AudioLock.Lock()
	if w.AudioTrack == nil {
		w.AudioLock.Unlock()
		return
	}

	w.AudioPacketsWritten += 1
	w.AudioTimestamp = uint32(int64(w.AudioTimestamp) + packet.TimeDiff)
	audioTrack := w.AudioTrack
	w.AudioLock.Unlock()

	if err := audioTrack.WriteRTP(packet.Packet, packet.Codec); err != nil {
		if errors.Is(err, io.ErrClosedPipe) {
			log.Println("WHEPSession.SendAudioPacket.ConnectionDropped")
			w.Close()
		} else {
			log.Println("WHEPSession.SendAudioPacket.Error", err)
		}
	}
}

// Sends provided video packet to the WHEP session
func (w *WHEPSession) SendVideoPacket(packet codecs.TrackPacket) {
	if w.IsSessionClosed.Load() {
		return
	}

	if w.IsWaitingForKeyframe.Load() {
		if !packet.IsKeyframe {
			w.SendPLI()
			return
		}

		w.IsWaitingForKeyframe.Store(false)
	}

	w.VideoLock.Lock()
	w.VideoBytesWritten += len(packet.Packet.Payload)
	w.VideoPacketsWritten += 1
	w.VideoSequenceNumber = uint16(w.VideoSequenceNumber) + uint16(packet.SequenceDiff)
	w.VideoTimestamp = uint32(int64(w.VideoTimestamp) + packet.TimeDiff)
	w.updateVideoBitrateLocked(time.Now())
	videoSequenceNumber := w.VideoSequenceNumber
	videoTimestamp := w.VideoTimestamp
	videoTrack := w.VideoTrack
	w.VideoLock.Unlock()

	if videoTrack == nil {
		return
	}

	packet.Packet.SequenceNumber = videoSequenceNumber
	packet.Packet.Timestamp = videoTimestamp

	if err := videoTrack.WriteRTP(packet.Packet, packet.Codec); err != nil {
		w.VideoPacketsDropped.Add(1)

		if errors.Is(err, io.ErrClosedPipe) {
			log.Println("WHEPSession.SendVideoPacket.ConnectionDropped")
			w.Close()
		} else {
			log.Println("WHEPSession.SendVideoPacket.Error", err)
		}
	}
}
