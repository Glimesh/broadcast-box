package track

import (
	"errors"
	"io"
	"log"
	"math"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	pionCodecs "github.com/pion/rtp/codecs"
)

func VideoWriter(remoteTrack *webrtc.TrackRemote, stream *session.WhipSession, peerConnection *webrtc.PeerConnection) {
	id := remoteTrack.RID()

	if id == "" {
		id = codecs.VideoTrackLabelDefault
	}

	codec := codecs.GetVideoTrackCodec(remoteTrack.Codec().MimeType)
	track, err := AddVideoTrack(stream, id, codec, &stream.WhepSessionsLock)
	if err != nil {
		log.Println("VideoWriter.AddTrack.Error:", err)
		return
	}
	track.Priority = getPrioritizedStreamingLayer(id, peerConnection.CurrentRemoteDescription().SDP)

	stream.OnTrackChan <- struct{}{}

	go subscribeStreamChannels(remoteTrack, stream, peerConnection)

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
			log.Println("video_writer", err)
		}

		if err = rtpPkt.Unmarshal(rtpBuf[:rtpRead]); err != nil {
			log.Println("video_writer", err)
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

		stream.WhepSessionsLock.RLock()
		for whepSession := range stream.WhepSessions {
			stream.WhepSessions[whepSession].SendVideoPacket(
				rtpPkt,
				id,
				timeDiff,
				sequenceDiff,
				codec,
				isKeyframe)

		}
		stream.WhepSessionsLock.RUnlock()
	}
}

const (
	naluTypeBitmask = 0x1f

	idrNALUType = 5
	spsNALUType = 7
	ppsNALUType = 8
)

func isPacketKeyframe(pkt *rtp.Packet, codec int, depacketizer rtp.Depacketizer) bool {
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

func subscribeStreamChannels(remoteTrack *webrtc.TrackRemote, stream *session.WhipSession, peerConnection *webrtc.PeerConnection) {
	for {
		{
			select {
			case <-stream.ActiveContext.Done():
				return
			case <-stream.PliChan:
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
