package session

import (
	"errors"
	"io"
	"log"

	"github.com/pion/rtp"
)

// Sends provided audio packet to the Whep session
func (session *WhepSession) SendAudioPacket(rtpPkt *rtp.Packet, layer string, timeDiff int64, sequenceDiff int, codec int) {
	currentLayer := session.AudioLayerCurrent.Load()

	if currentLayer == "" {
		session.AudioLayerCurrent.Store(layer)
	} else if layer != currentLayer {
		return
	}

	session.AudioLock.Lock()
	session.AudioPacketsWritten += 1
	session.AudioSequenceNumber = uint16(session.AudioSequenceNumber) + uint16(sequenceDiff)
	session.AudioTimestamp = uint32(int64(session.AudioTimestamp) + timeDiff)

	rtpPkt.SequenceNumber = session.AudioSequenceNumber
	rtpPkt.Timestamp = session.AudioTimestamp
	session.AudioLock.Unlock()

	if err := session.AudioTrack.WriteRTP(rtpPkt, codec); err != nil && !errors.Is(err, io.ErrClosedPipe) {
		log.Println("WHEP.SendAudioPacket.Error", err)
	}
}

// Sends provided video packet to the Whep session
func (session *WhepSession) SendVideoPacket(rtpPkt *rtp.Packet, layer string, timeDiff int64, sequenceDiff int, codec int, isKeyframe bool) {
	session.VideoLock.Lock()
	currentLayer := session.VideoLayerCurrent.Load()

	if currentLayer == "" {
		session.VideoLayerCurrent.Store(layer)
	} else if layer != currentLayer {
		return
	} else if session.IsWaitingForKeyframe.Load() {
		if !isKeyframe {
			return
		}

		session.IsWaitingForKeyframe.Store(false)
	}

	session.VideoPacketsWritten += 1
	session.VideoSequenceNumber = uint16(session.VideoSequenceNumber) + uint16(sequenceDiff)
	session.VideoTimestamp = uint32(int64(session.VideoTimestamp) + timeDiff)

	rtpPkt.SequenceNumber = session.VideoSequenceNumber
	rtpPkt.Timestamp = session.VideoTimestamp
	session.VideoLock.Unlock()

	if err := session.VideoTrack.WriteRTP(rtpPkt, codec); err != nil && !errors.Is(err, io.ErrClosedPipe) {
		log.Println("WHEP.SendVideoPacket.Error", err)
	}
}
