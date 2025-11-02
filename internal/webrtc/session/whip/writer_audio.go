package whip

import (
	"errors"
	"io"
	"log"
	"math"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/session/whep"
)

func (whipSession *WhipSession) AudioWriter(remoteTrack *webrtc.TrackRemote, peerConnection *webrtc.PeerConnection) {
	id := remoteTrack.RID()

	if id == "" {
		id = codecs.AudioTrackLabelDefault
	}

	codec := codecs.GetAudioTrackCodec(remoteTrack.Codec().MimeType)
	track, err := whipSession.AddAudioTrack(id, codec)
	if err != nil {
		log.Println("AudioWriter.AddTrack.Error:", err)
		return
	}
	track.Priority = whipSession.getPrioritizedStreamingLayer(id, peerConnection.CurrentRemoteDescription().SDP)

	rtpBuf := make([]byte, 1500)
	rtpPkt := &rtp.Packet{}

	lastTimestamp := uint32(0)
	lastTimestampSet := false

	lastSequenceNumber := uint16(0)
	lastSequenceNumberSet := false

	for {
		sessionsAny := whipSession.WhepSessionsSnapshot.Load()
		if sessionsAny == nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		sessions := sessionsAny.(map[string]*whep.WhepSession)

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

		for _, whepSession := range sessions {

			select {
			case whepSession.AudioChannel <- codecs.TrackPacket{
				Layer: id,
				Packet: &rtp.Packet{
					Header:  rtpPkt.Header,
					Payload: append([]byte(nil), rtpPkt.Payload...),
				},
				Codec:        codec,
				TimeDiff:     timeDiff,
				SequenceDiff: sequenceDiff,
			}:
			default:
				// Drop packet
			}

		}

	}
}
