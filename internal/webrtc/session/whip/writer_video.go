package whip

import (
	"errors"
	"io"
	"log"
	"math"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	pionCodecs "github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
)

func (whipSession *WhipSession) VideoWriter(remoteTrack *webrtc.TrackRemote, peerConnection *webrtc.PeerConnection) {
	id := remoteTrack.RID()

	if id == "" {
		id = codecs.VideoTrackLabelDefault
	}

	codec := codecs.GetVideoTrackCodec(remoteTrack.Codec().MimeType)
	track, err := whipSession.AddVideoTrack(id, codec)
	if err != nil {
		log.Println("VideoWriter.AddTrack.Error:", err)
		return
	}
	track.Priority = whipSession.getPrioritizedStreamingLayer(id, peerConnection.CurrentRemoteDescription().SDP)

	go whipStreamVideoWriterChannels(remoteTrack, whipSession, peerConnection)

	rtpBuf := make([]byte, 1500)
	rtpPkt := &rtp.Packet{}

	var depacketizer rtp.Depacketizer
	switch codec {
	case codecs.VideoTrackCodecH264:
		depacketizer = &pionCodecs.H264Packet{}
	case codecs.VideoTrackCodecVP8:
		depacketizer = &pionCodecs.VP8Packet{}
	case codecs.VideoTrackCodecVP9:
		depacketizer = &pionCodecs.VP9Packet{}
	}

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
			log.Println("VideoWriter.RtpPkt.Unmarshal.Error", err)
		}

		if err = rtpPkt.Unmarshal(rtpBuf[:rtpRead]); err != nil {
			log.Println("VideoWriter.RtpPkt.Unmarshal.Error", err)
			return
		}

		track.PacketsReceived.Add(1)

		isKeyframe := false
		if codec == codecs.VideoTrackCodecH264 {
			isKeyframe := isPacketKeyframe(rtpPkt, codec, depacketizer)
			if isKeyframe {
				track.LastKeyFrame.Store(time.Now())
			}
		}

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
			sequenceDiff = 0
		case sequenceDiff < -(math.MaxUint16 / 10):
			sequenceDiff += (math.MaxUint16 + 1)
		}

		lastTimestamp = rtpPkt.Timestamp
		lastSequenceNumber = rtpPkt.SequenceNumber

		whipSession.WhepSessionsLock.RLock()
		for _, whepSession := range whipSession.WhepSessions {
			whepSession.SendVideoPacket(
				rtpPkt,
				id,
				timeDiff,
				sequenceDiff,
				codec,
				isKeyframe)
		}
		whipSession.WhepSessionsLock.RUnlock()
	}
}

const (
	naluTypeBitmask = 0x1f

	idrNALUType = 5
	spsNALUType = 7
	ppsNALUType = 8
)

func isPacketKeyframe(pkt *rtp.Packet, codec codecs.TrackCodeType, depacketizer rtp.Depacketizer) bool {
	if codec == codecs.VideoTrackCodecH264 {
		nalu, err := depacketizer.Unmarshal(pkt.Payload)

		if err != nil || len(nalu) < 6 {
			return false
		}

		firstNaluType := nalu[4] & naluTypeBitmask
		return firstNaluType == idrNALUType || firstNaluType == spsNALUType || firstNaluType == ppsNALUType
	}

	return true
}

// Triggers a request for a new key frame to be sent to the client
func whipStreamVideoWriterChannels(remoteTrack *webrtc.TrackRemote, whipSession *WhipSession, peerConnection *webrtc.PeerConnection) {

	whipSession.StatusLock.RLock()
	whipActiveContext := whipSession.ActiveContext
	whipSession.StatusLock.RUnlock()

	for {
		{
			select {
			case <-whipActiveContext.Done():
				return
			case <-whipSession.PacketLossIndicationChannel:
				if sendError := peerConnection.WriteRTCP([]rtcp.Packet{
					&rtcp.PictureLossIndication{
						MediaSSRC: uint32(remoteTrack.SSRC()),
					},
				}); sendError != nil {
					return
				}
			}
		}
	}
}
