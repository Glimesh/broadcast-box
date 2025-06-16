package webrtc

import (
	"errors"
	"io"
	"log"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
)

func audioWriter(remoteTrack *webrtc.TrackRemote, stream *stream) {
	rtpBuf := make([]byte, 1500)
	for {
		rtpRead, _, err := remoteTrack.Read(rtpBuf)
		switch {
		case errors.Is(err, io.EOF):
			return
		case err != nil:
			log.Println(err)
			return
		}

		stream.audioPacketsReceived.Add(1)
		if _, writeErr := stream.audioTrack.Write(rtpBuf[:rtpRead]); writeErr != nil && !errors.Is(writeErr, io.ErrClosedPipe) {
			log.Println(writeErr)
			return
		}
	}
}

func videoWriter(remoteTrack *webrtc.TrackRemote, stream *stream, peerConnection *webrtc.PeerConnection, s *stream, sessionId string) {
	id := remoteTrack.RID()
	if id == "" {
		id = videoTrackLabelDefault
	}

	videoTrack, err := addTrack(s, id, sessionId)
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		for {
			select {
			case <-stream.whipActiveContext.Done():
				return
			case <-stream.pliChan:
				if sendErr := peerConnection.WriteRTCP([]rtcp.Packet{
					&rtcp.PictureLossIndication{
						MediaSSRC: uint32(remoteTrack.SSRC()),
					},
				}); sendErr != nil {
					return
				}
			}
		}
	}()

	rtpBuf := make([]byte, 1500)
	rtpPkt := &rtp.Packet{}
	codec := getVideoTrackCodec(remoteTrack.Codec().MimeType)

	var depacketizer rtp.Depacketizer
	switch codec {
	case videoTrackCodecH264:
		depacketizer = &codecs.H264Packet{}
	case videoTrackCodecVP8:
		depacketizer = &codecs.VP8Packet{}
	case videoTrackCodecVP9:
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
			log.Println(err)
			return
		}

		if err = rtpPkt.Unmarshal(rtpBuf[:rtpRead]); err != nil {
			log.Println(err)
			return
		}

		videoTrack.packetsReceived.Add(1)

		// Keyframe detection has only been implemented for H264
		isKeyframe := isKeyframe(rtpPkt, codec, depacketizer)
		if isKeyframe && codec == videoTrackCodecH264 {
			videoTrack.lastKeyFrameSeen.Store(time.Now())
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

		s.whepSessionsLock.RLock()
		for i := range s.whepSessions {
			s.whepSessions[i].sendVideoPacket(rtpPkt, id, timeDiff, sequenceDiff, codec, isKeyframe)
		}
		s.whepSessionsLock.RUnlock()

	}
}

func WHIP(offer, streamKey string) (string, error) {
	maybePrintOfferAnswer(offer, true)

	whipSessionId := uuid.New().String()

	peerConnection, err := newPeerConnection(apiWhip)
	if err != nil {
		return "", err
	}

	streamMapLock.Lock()
	defer streamMapLock.Unlock()
	stream, err := getStream(streamKey, whipSessionId)
	if err != nil {
		return "", err
	}

	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
		if strings.HasPrefix(remoteTrack.Codec().MimeType, "audio") {
			audioWriter(remoteTrack, stream)
		} else {
			videoWriter(remoteTrack, stream, peerConnection, stream, whipSessionId)

		}
	})

	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		if i == webrtc.ICEConnectionStateFailed || i == webrtc.ICEConnectionStateClosed {
			if err := peerConnection.Close(); err != nil {
				log.Println(err)
			}
			peerConnectionDisconnected(true, streamKey, whipSessionId)
		}
	})

	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  string(offer),
		Type: webrtc.SDPTypeOffer,
	}); err != nil {
		return "", err
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)

	if err != nil {
		return "", err
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", err
	}

	<-gatherComplete
	return maybePrintOfferAnswer(appendAnswer(peerConnection.LocalDescription().SDP), false), nil
}
