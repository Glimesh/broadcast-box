package whip

import (
	"errors"
	"io"
	"log"
	"math"
	"strings"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
	"github.com/pion/rtp"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"

	pionCodecs "github.com/pion/rtp/codecs"
)

func (w *WHIPSession) audioWriter(remoteTrack *webrtc.TrackRemote, streamKey string) {
	id := remoteTrack.RID()

	if id == "" {
		id = codecs.AudioTrackLabelDefault
	}

	codec := codecs.GetAudioTrackCodec(remoteTrack.Codec().MimeType)
	track, err := w.addAudioTrack(id, streamKey, codec)
	if err != nil {
		log.Println("AudioWriter.AddTrack.Error:", err)
		return
	}

	rtpPkt := &rtp.Packet{}
	rtpBuf := make([]byte, 1500)
	for {
		rtpRead, _, err := remoteTrack.Read(rtpBuf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("WHIPSession.AudioWriter.RtpPkt.EndOfStream")
				return
			} else {
				log.Println("WHIPSession.AudioWriter.RtpPkt.Err", err)
			}
		}

		track.PacketsReceived.Add(1)

		err = rtpPkt.Unmarshal(rtpBuf[:rtpRead])
		if err != nil {
			log.Println("WHIPSession.AudioWriter.RtpPkt.Error", err)
			continue
		}

		var sessions map[string]*whep.WHEPSession
		if sessionsAny := w.WHEPSessionsSnapshot.Load(); sessionsAny != nil {
			sessions = sessionsAny.(map[string]*whep.WHEPSession)
		}

		packet := codecs.TrackPacket{
			Layer:  id,
			Packet: rtpPkt,
			Codec:  codec,
		}

		for _, whepSession := range sessions {
			whepSession.SendAudioPacket(packet)
		}
	}
}

func (w *WHIPSession) videoWriter(remoteTrack *webrtc.TrackRemote, streamKey string, peerConnection *webrtc.PeerConnection) {
	id := remoteTrack.RID()

	if id == "" {
		id = codecs.VideoTrackLabelDefault
	}

	codec := codecs.GetVideoTrackCodec(remoteTrack.Codec().MimeType)
	track, err := w.addVideoTrack(id, streamKey, codec)
	if err != nil {
		log.Println("WHIPSession.VideoWriter.AddTrack.Error:", err)
		return
	}
	track.Priority = w.getPrioritizedStreamingLayer(id, peerConnection.CurrentRemoteDescription().SDP)
	track.MediaSSRC.Store(uint32(remoteTrack.SSRC()))

	var depacketizer rtp.Depacketizer
	switch codec {
	case codecs.VideoTrackCodecH264:
		depacketizer = &pionCodecs.H264Packet{}
	case codecs.VideoTrackCodecH265:
		depacketizer = &pionCodecs.H265Depacketizer{}
	case codecs.VideoTrackCodecVP8:
		depacketizer = &pionCodecs.VP8Packet{}
	case codecs.VideoTrackCodecVP9:
		depacketizer = &pionCodecs.VP9Packet{}
	case codecs.VideoTrackCodecAV1:
		depacketizer = &pionCodecs.AV1Depacketizer{}
	}

	if depacketizer == nil {
		log.Println("WHIPSession.VideoWriter.Depacketizer: No depacketizer was found for codec", codec)
	}

	lastTimestamp := uint32(0)
	lastTimestampSet := false

	lastSequenceNumber := uint16(0)
	lastSequenceNumberSet := false

	bitrateWindowStart := time.Now()
	bitrateWindowBytes := uint64(0)

	rtpPkt := &rtp.Packet{}
	pktBuf := make([]byte, 1500)
	for {
		rtpRead, _, err := remoteTrack.Read(pktBuf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("WHIPSession.VideoWriter.RtpPkt.EndOfStream")
				w.notifyClosed()
				return
			} else {
				log.Println("WHIPSession.VideoWriter.RtpPkt.Err", err)
			}
		}

		if rtpRead == 0 {
			continue
		}

		err = rtpPkt.Unmarshal(pktBuf[:rtpRead])
		if err != nil {
			log.Println("WHIPSession.VideoWriter.RtpPkt.Unmarshal", err)
			continue
		}

		rtpPkt.Extension = false
		rtpPkt.Extensions = nil

		track.PacketsReceived.Add(1)
		bitrateWindowBytes += uint64(rtpRead)

		isKeyframe := isPacketKeyframe(rtpPkt, codec, depacketizer)
		if isKeyframe {
			track.LastKeyFrame.Store(time.Now())
		}

		now := time.Now()
		if elapsed := now.Sub(bitrateWindowStart); elapsed >= time.Second {
			track.Bitrate.Store(uint64(float64(bitrateWindowBytes) / elapsed.Seconds()))
			bitrateWindowStart = now
			bitrateWindowBytes = 0
		}

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

		var sessions map[string]*whep.WHEPSession
		if sessionsAny := w.WHEPSessionsSnapshot.Load(); sessionsAny != nil {
			sessions = sessionsAny.(map[string]*whep.WHEPSession)
		}

		for _, whepSession := range sessions {
			if whepSession.GetVideoLayerOrDefault(id, track.Priority) != id {
				continue
			}

			whepSession.SendVideoPacket(codecs.TrackPacket{
				Layer:        id,
				Packet:       rtpPkt,
				Codec:        codec,
				IsKeyframe:   isKeyframe,
				TimeDiff:     timeDiff,
				SequenceDiff: sequenceDiff,
			})
		}
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

// Helper function for getting the simulcast order and using as priority for consumers
// This example will order from left to right with highest to lowest priority
// a=simulcast:send High,Mid,Low
func (w *WHIPSession) getPrioritizedStreamingLayer(layer string, sdpDescription string) int {
	var sessionDescription sdp.SessionDescription
	err := sessionDescription.Unmarshal([]byte(sdpDescription))
	if err != nil {
		log.Println("Track.getPrioritizedStreamingLayer Error: (Layer "+layer+")", err)
		return 100
	}

	var priority = 1
	for _, description := range sessionDescription.MediaDescriptions {
		for _, attribute := range description.Attributes {
			if attribute.Key == "simulcast" && strings.HasPrefix(attribute.Value, "send ") {
				layers := strings.TrimPrefix(attribute.Value, "send")
				log.Println("WHIPSession.VideoWriter.TrackPriority:", layers)
				for simulcastLayer := range strings.SplitSeq(strings.TrimSpace(layers), ";") {
					if simulcastLayer != "" && strings.EqualFold(simulcastLayer, layer) {
						log.Println("WHIPSession.VideoWriter.TrackPriority:", layer)
						return priority
					} else {
						priority++
					}
				}
			}
		}
	}

	return 100
}
