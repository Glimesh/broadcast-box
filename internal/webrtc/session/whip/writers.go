package whip

import (
	"errors"
	"io"
	"log"
	"math"
	"strings"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/session/whep"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	pionCodecs "github.com/pion/rtp/codecs"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"
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

	rtpPkt := &rtp.Packet{}
	rtpBuf := make([]byte, 1500)
	for {
		rtpRead, _, err := remoteTrack.Read(rtpBuf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("WhipSession.AudioWriter.RtpPkt.EndOfStream")
				return
			}
		}

		// TODO
		// Consider creating the snapshot in upper scope and update every 1 sec instead of loading every iteration
		sessionsAny := whipSession.WhepSessionsSnapshot.Load()
		if sessionsAny == nil {
			continue
		}

		sessions := sessionsAny.(map[string]*whep.WhepSession)

		track.PacketsReceived.Add(1)

		err = rtpPkt.Unmarshal(rtpBuf[:rtpRead])
		if err != nil {
			log.Println("WhipSession.AudioWriter.RtpPkt.Error", err)
			continue
		}

		packet := codecs.TrackPacket{
			Layer:  id,
			Packet: rtpPkt,
			Codec:  codec,
		}

		for _, whepSession := range sessions {
			if whepSession.AudioLayerCurrent.Load() == id {
				whepSession.SendAudioPacket(packet)
			}
		}
	}
}

func (whipSession *WhipSession) VideoWriter(remoteTrack *webrtc.TrackRemote, peerConnection *webrtc.PeerConnection) {
	id := remoteTrack.RID()

	if id == "" {
		id = codecs.VideoTrackLabelDefault
	}

	codec := codecs.GetVideoTrackCodec(remoteTrack.Codec().MimeType)
	track, err := whipSession.AddVideoTrack(id, codec)
	if err != nil {
		log.Println("WhipSession.VideoWriter.AddTrack.Error:", err)
		return
	}
	track.Priority = whipSession.getPrioritizedStreamingLayer(id, peerConnection.CurrentRemoteDescription().SDP)

	go whipStreamVideoWriterChannels(remoteTrack, whipSession, peerConnection)

	var depacketizer rtp.Depacketizer
	switch codec {
	case codecs.VideoTrackCodecH264:
		depacketizer = &pionCodecs.H264Packet{}
	case codecs.VideoTrackCodecH265:
		depacketizer = &pionCodecs.H265Packet{}
	case codecs.VideoTrackCodecVP8:
		depacketizer = &pionCodecs.VP8Packet{}
	case codecs.VideoTrackCodecVP9:
		depacketizer = &pionCodecs.VP9Packet{}
	}

	if depacketizer == nil {
		log.Println("WhipSession.VideoWriter.Depacketizer: No depacketizer was found for codec", codec)
	}

	lastTimestamp := uint32(0)
	lastTimestampSet := false

	lastSequenceNumber := uint16(0)
	lastSequenceNumberSet := false

	for {
		rtpPkt, _, err := remoteTrack.ReadRTP()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("WhipSession.VideoWriter.RtpPkt.EndOfStream")
				return
			}
		}

		sessionsAny := whipSession.WhepSessionsSnapshot.Load()

		if sessionsAny == nil {
			continue
		}

		sessions := sessionsAny.(map[string]*whep.WhepSession)

		// Consider using a variable that occassionaly updates the atomic instead
		track.PacketsReceived.Add(1)

		isKeyframe := false
		if codec == codecs.VideoTrackCodecH264 {
			isKeyframe = isPacketKeyframe(rtpPkt, codec, depacketizer)
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

		packet := codecs.TrackPacket{
			Layer:        id,
			Packet:       rtpPkt,
			Codec:        codec,
			IsKeyframe:   isKeyframe,
			TimeDiff:     timeDiff,
			SequenceDiff: sequenceDiff,
		}

		for _, whepSession := range sessions {
			if whepSession.VideoLayerCurrent.Load() == id {
				whepSession.SendVideoPacket(packet)
			}
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

// Triggers a request for a new key frame
func whipStreamVideoWriterChannels(remoteTrack *webrtc.TrackRemote, whipSession *WhipSession, peerConnection *webrtc.PeerConnection) {
	for {
		{
			select {
			case <-whipSession.ActiveContext.Done():
				return
			case <-whipSession.PacketLossIndicationChannel:
				log.Println("WhipSession.WhipStreamVideoWriterChannels.Trigger.PLI")
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

// Helper function for getting the simulcast order and using as priority for consumers
// This example will order from left to right with highest to lowest priority
// a=simulcast:send High,Mid,Low
func (whipSession *WhipSession) getPrioritizedStreamingLayer(layer string, sdpDescription string) int {
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
				log.Println("WhipSession.VideoWriter.TrackPriority:", layers)
				for simulcastLayer := range strings.SplitSeq(strings.TrimSpace(layers), ";") {
					if simulcastLayer != "" && strings.EqualFold(simulcastLayer, layer) {
						log.Println("WhipSession.VideoWriter.TrackPriority:", layer)
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
