package whip

import (
	"errors"
	"io"
	"log"
	"math"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	broadcastCodecs "github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/stream"
)

func VideoWriter(remoteTrack *webrtc.TrackRemote, session *stream.WhipSession, peerConnection *webrtc.PeerConnection) {
	id := remoteTrack.RID()

	if id == "" {
		session.WhepSessionsLock.RLock()
		var names []string
		for _, track := range session.VideoTracks {
			names = append(names, track.Rid)
		}
		session.WhepSessionsLock.RUnlock()

		id = NextAvailableName(broadcastCodecs.VideoTrackLabelDefault, names)
	}

	track, err := stream.AddVideoTrack(session, id, &session.WhepSessionsLock)

	if err != nil {
		log.Println("VideoWriter.AddTrack.Error:", err)
		return
	}

	go func() {
		for {
			{
				select {
				case <-session.ActiveContext.Done():
					return
				case <-session.PliChan:
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
	}()

	rtpBuf := make([]byte, 1500)
	rtpPkt := &rtp.Packet{}
	codec := broadcastCodecs.GetVideoTrackCodec(remoteTrack.Codec().MimeType)

	var depacketizer rtp.Depacketizer
	switch codec {
	case broadcastCodecs.VideoTrackCodecH264:
		depacketizer = &codecs.H264Packet{}
	case broadcastCodecs.VideoTrackCodecVP8:
		depacketizer = &codecs.VP8Packet{}
	case broadcastCodecs.VideoTrackCodecVP9:
		depacketizer = &codecs.VP9Packet{}
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
			log.Println("video_writer", err)
		}

		if err = rtpPkt.Unmarshal(rtpBuf[:rtpRead]); err != nil {
			log.Println("video_writer", err)
			return
		}

		track.PacketsReceived.Add(1)

		isKeyframe := false
		if codec == broadcastCodecs.VideoTrackCodecH264 {
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

		session.WhepSessionsLock.RLock()
		for whepSession := range session.WhepSessions {
			session.WhepSessions[whepSession].SendVideoPacket(
				rtpPkt,
				id,
				timeDiff,
				sequenceDiff,
				codec,
				isKeyframe)

		}
		session.WhepSessionsLock.RUnlock()
	}
}

const (
	naluTypeBitmask = 0x1f

	idrNALUType = 5
	spsNALUType = 7
	ppsNALUType = 8
)

func isPacketKeyframe(pkt *rtp.Packet, codec int, depacketizer rtp.Depacketizer) bool {
	if codec == broadcastCodecs.VideoTrackCodecH264 {
		nalu, err := depacketizer.Unmarshal(pkt.Payload)

		if err != nil || len(nalu) < 6 {
			return false
		}

		firstNaluType := nalu[4] & naluTypeBitmask
		return firstNaluType == idrNALUType || firstNaluType == spsNALUType || firstNaluType == ppsNALUType
	}

	return true
}
