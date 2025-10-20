package whep

import (
	"errors"
	"io"
	"log"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/rtp"
)

// Sends provided audio packet to the Whep session
func (whepSession *WhepSession) SendAudioPacket(rtpPkt *rtp.Packet, layer string, timeDiff int64, sequenceDiff int, codec codecs.TrackCodeType) {
	whepSession.AudioLock.RLock()
	audioTrack := whepSession.AudioTrack
	whepSession.AudioLock.RUnlock()

	if whepSession.IsSessionClosed.Load() || audioTrack == nil {
		return
	}

	currentLayer := whepSession.AudioLayerCurrent.Load()

	if currentLayer == "" {
		whepSession.AudioLayerCurrent.Store(layer)
	} else if layer != currentLayer {
		return
	}

	// Convert to WhepSession Function
	whepSession.AudioLock.Lock()
	whepSession.AudioPacketsWritten += 1
	whepSession.AudioSequenceNumber = uint16(whepSession.AudioSequenceNumber) + uint16(sequenceDiff)
	whepSession.AudioTimestamp = uint32(int64(whepSession.AudioTimestamp) + timeDiff)

	rtpPkt.SequenceNumber = whepSession.AudioSequenceNumber
	rtpPkt.Timestamp = whepSession.AudioTimestamp
	whepSession.AudioLock.Unlock()

	if err := whepSession.AudioTrack.WriteRTP(rtpPkt, codec); err != nil {
		if errors.Is(err, io.ErrClosedPipe) {
			log.Println("WhepSession.SendAudioPacket.ConnectionDropped")
			whepSession.Close()
		} else {
			log.Println("WhepSession.SendAudioPacket.Error", err)
		}
	}
}

// Sends provided video packet to the Whep session
func (whepSession *WhepSession) SendVideoPacket(rtpPkt *rtp.Packet, layer string, timeDiff int64, sequenceDiff int, codec codecs.TrackCodeType, isKeyframe bool) {
	whepSession.VideoLock.RLock()
	videoTrack := whepSession.VideoTrack
	whepSession.VideoLock.RUnlock()

	if whepSession.IsSessionClosed.Load() || videoTrack == nil {
		return
	}

	currentLayer := whepSession.VideoLayerCurrent.Load()

	if currentLayer == "" {
		whepSession.VideoLayerCurrent.Store(layer)
	} else if layer != currentLayer {
		return
	} else if whepSession.IsWaitingForKeyframe.Load() {
		if !isKeyframe {
			return
		}

		whepSession.IsWaitingForKeyframe.Store(false)
	}

	// Convert to WhepSession Function
	whepSession.VideoLock.Lock()
	whepSession.VideoPacketsWritten += 1
	whepSession.VideoSequenceNumber = uint16(whepSession.VideoSequenceNumber) + uint16(sequenceDiff)
	whepSession.VideoTimestamp = uint32(int64(whepSession.VideoTimestamp) + timeDiff)

	rtpPkt.SequenceNumber = whepSession.VideoSequenceNumber
	rtpPkt.Timestamp = whepSession.VideoTimestamp
	whepSession.VideoLock.Unlock()

	if err := whepSession.VideoTrack.WriteRTP(rtpPkt, codec); err != nil {
		if errors.Is(err, io.ErrClosedPipe) {
			log.Println("WhepSession.SendVideoPacket.ConnectionDropped")
			whepSession.Close()
		} else {
			log.Println("WhepSession.SendVideoPacket.Error", err)
		}
	}
}
