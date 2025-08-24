package track

import (
	"errors"
	"io"
	"log"
	"math"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
)

func AudioWriter(remoteTrack *webrtc.TrackRemote, stream *session.WhipSession, peerConnection *webrtc.PeerConnection) {
	id := remoteTrack.RID()

	if id == "" {
		id = codecs.AudioTrackLabelDefault
	}

	codec := codecs.GetAudioTrackCodec(remoteTrack.Codec().MimeType)
	track, err := AddAudioTrack(stream, id, codec, &stream.WhepSessionsLock)
	if err != nil {
		log.Println("AudioWriter.AddTrack.Error:", err)
		return
	}
	track.Priority = getPrioritizedStreamingLayer(id, peerConnection.CurrentRemoteDescription().SDP)

	stream.OnTrackChan <- struct{}{}

	rtpBuf := make([]byte, 1500)
	rtpPkt := &rtp.Packet{}

	lastTimestamp := uint32(0)
	lastTimestampSet := false

	lastSequenceNumber := uint16(0)
	lastSequenceNumberSet := false

	for {
		rtpRead, _, err := remoteTrack.Read(rtpBuf)

		switch {
		case errors.Is(err, io.EOF):
			return
		case err != nil:
			log.Println(err)
			return
		}

		if err = rtpPkt.Unmarshal(rtpBuf[:rtpRead]); err != nil {
			log.Println(err)
			return
		}

		track.PacketsReceived.Add(1)

		rtpPkt.Extension = false
		rtpPkt.Extensions = nil

		timeDiff := int64(rtpPkt.Timestamp) - int64(lastTimestamp)
		switch {
		case !lastTimestampSet:
			timeDiff = 0
			lastTimestampSet = true
		case timeDiff < -(math.MaxUint32 / 10):
			timeDiff += (math.MaxUint32 + 1)
		}

		sequenceDiff := int(rtpPkt.SequenceNumber) - int(lastSequenceNumber)
		switch {
		case !lastSequenceNumberSet:
			lastSequenceNumberSet = true
		case sequenceDiff < -(math.MaxUint16 / 10):
			sequenceDiff += (math.MaxUint32 + 1)
		}

		lastTimestamp = rtpPkt.Timestamp
		lastSequenceNumber = rtpPkt.SequenceNumber

		stream.WhepSessionsLock.RLock()
		for whepSession := range stream.WhepSessions {
			stream.WhepSessions[whepSession].SendAudioPacket(
				rtpPkt,
				id,
				timeDiff,
				sequenceDiff,
				codec)
		}
		stream.WhepSessionsLock.RUnlock()
	}
}
