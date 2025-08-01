package stream

import (
	"errors"
	"io"
	"log"
	"sync/atomic"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type (
	WhepSession struct {
		SessionId            string
		IsWaitingForKeyframe atomic.Bool

		VideoTrack          *TrackMultiCodec
		VideoTimestamp      uint32
		VideoPacketsWritten uint64
		VideoSequenceNumber uint16
		VideoLayerCurrent   atomic.Value

		AudioTrack          *TrackMultiCodec
		AudioTimestamp      uint32
		AudioPacketsWritten uint64
		AudioSequenceNumber uint16
		AudioLayerCurrent   atomic.Value

		OnTrackHandler func(stream *WhipSession) func(*webrtc.TrackRemote, *webrtc.RTPReceiver)
	}
)

func (session *WhepSession) SendAudioPacket(rtpPkt *rtp.Packet, layer string, timeDiff int64, sequenceDiff int, codec int) {
	currentLayer := session.AudioLayerCurrent.Load()

	if currentLayer == "" {
		session.AudioLayerCurrent.Store(layer)
	} else if layer != currentLayer {
		return
	}

	session.AudioPacketsWritten += 1
	session.AudioSequenceNumber = uint16(session.AudioSequenceNumber) + uint16(sequenceDiff)
	session.AudioTimestamp = uint32(int64(session.AudioTimestamp) + timeDiff)

	rtpPkt.SequenceNumber = session.AudioSequenceNumber
	rtpPkt.Timestamp = session.AudioTimestamp

	if err := session.AudioTrack.WriteRTP(rtpPkt, codec); err != nil && !errors.Is(err, io.ErrClosedPipe) {
		log.Println("SendAudioPacket.Error", err)
	}
}

func (session *WhepSession) SendVideoPacket(rtpPkt *rtp.Packet, layer string, timeDiff int64, sequenceDiff int, codec int, isKeyframe bool) {
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

	if err := session.VideoTrack.WriteRTP(rtpPkt, codec); err != nil && !errors.Is(err, io.ErrClosedPipe) {
		log.Println("SendVideoPacket.Error", err)
	}
}
