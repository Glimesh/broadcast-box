package codecs

import (
	"log"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type TrackPacket struct {
	Layer        string
	Packet       *rtp.Packet
	TimeDiff     int64
	SequenceDiff int
	Codec        TrackCodeType
	IsKeyframe   bool
}

type TrackMultiCodec struct {
	id         string
	rid        string
	streamID   string
	kind       webrtc.RTPCodecType
	codec      TrackCodeType
	errorCount int

	ssrc        webrtc.SSRC
	writeStream webrtc.TrackLocalWriter

	payloadTypeH264 uint8
	payloadTypeH265 uint8
	payloadTypeVP8  uint8
	payloadTypeVP9  uint8
	payloadTypeAV1  uint8
	payloadTypeOpus uint8

	currentPayloadType uint8
}

func (t *TrackMultiCodec) ID() string                { return t.id }
func (t *TrackMultiCodec) RID() string               { return t.rid }
func (t *TrackMultiCodec) StreamID() string          { return t.streamID }
func (t *TrackMultiCodec) Kind() webrtc.RTPCodecType { return t.kind }

func CreateTrackMultiCodec(id string, rid string, streamID string, kind webrtc.RTPCodecType, codec TrackCodeType) *TrackMultiCodec {
	return &TrackMultiCodec{
		id:       id,
		rid:      rid,
		streamID: streamID,
		kind:     kind,
		codec:    codec,
	}
}

func (t *TrackMultiCodec) Bind(ctx webrtc.TrackLocalContext) (webrtc.RTPCodecParameters, error) {
	t.ssrc = ctx.SSRC()
	t.writeStream = ctx.WriteStream()

	var videoCodecParameters webrtc.RTPCodecParameters
	codecParameters := ctx.CodecParameters()
	for parameters := range codecParameters {
		switch GetAudioTrackCodec(codecParameters[parameters].MimeType) {
		case audioTrackCodecOpus:
			t.payloadTypeOpus = uint8(codecParameters[parameters].PayloadType)
			t.currentPayloadType = t.payloadTypeOpus
		}

		if t.payloadTypeOpus != 0 {
			log.Println("WHIPSession.TrackMultiCodec: Binding AudioTrack Type for", t.streamID, "-", t.currentPayloadType)

			t.kind = webrtc.RTPCodecTypeAudio
			return webrtc.RTPCodecParameters{
				PayloadType: codecParameters[parameters].PayloadType,
				RTPCodecCapability: webrtc.RTPCodecCapability{
					MimeType:     codecParameters[parameters].MimeType,
					RTCPFeedback: codecParameters[parameters].RTCPFeedback,
					ClockRate:    codecParameters[parameters].ClockRate,
					SDPFmtpLine:  codecParameters[parameters].SDPFmtpLine,
				},
			}, nil
		}

		switch GetVideoTrackCodec(codecParameters[parameters].MimeType) {
		case VideoTrackCodecH264:
			t.payloadTypeH264 = uint8(codecParameters[parameters].PayloadType)
			t.currentPayloadType = t.payloadTypeH264
			videoCodecParameters = codecParameters[parameters]

		case VideoTrackCodecH265:
			t.payloadTypeH265 = uint8(codecParameters[parameters].PayloadType)
			t.currentPayloadType = t.payloadTypeH265
			videoCodecParameters = codecParameters[parameters]

		case VideoTrackCodecVP8:
			t.payloadTypeVP8 = uint8(codecParameters[parameters].PayloadType)
			t.currentPayloadType = t.payloadTypeVP8
			videoCodecParameters = codecParameters[parameters]

		case VideoTrackCodecVP9:
			t.payloadTypeVP9 = uint8(codecParameters[parameters].PayloadType)
			t.currentPayloadType = t.payloadTypeVP9
			videoCodecParameters = codecParameters[parameters]

		case VideoTrackCodecAV1:
			t.payloadTypeAV1 = uint8(codecParameters[parameters].PayloadType)
			t.currentPayloadType = t.payloadTypeAV1
			videoCodecParameters = codecParameters[parameters]
		}
	}

	log.Println("WHEPSession.TrackMultiCodec: Binding VideoTrack Type for", t.streamID, "-", t.currentPayloadType)
	t.kind = webrtc.RTPCodecTypeVideo
	return webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     videoCodecParameters.MimeType,
			RTCPFeedback: videoCodecParameters.RTCPFeedback,
			ClockRate:    videoCodecParameters.ClockRate,
			SDPFmtpLine:  videoCodecParameters.SDPFmtpLine,
			Channels:     videoCodecParameters.Channels,
		},
	}, nil
}

func (t *TrackMultiCodec) Unbind(context webrtc.TrackLocalContext) error {
	return nil
}

func (t *TrackMultiCodec) WriteRTP(packet *rtp.Packet, codec TrackCodeType) error {
	packet.SSRC = uint32(t.ssrc)

	if codec != t.codec {
		log.Println("WHEPSession.TrackMultiCodec.WriteRTP: Setting Codec on", t.streamID, "(", t.RID(), ")", "from", t.codec, "to", codec)
		t.codec = codec

		switch t.codec {
		case VideoTrackCodecH264:
			t.currentPayloadType = t.payloadTypeH264
		case VideoTrackCodecH265:
			t.currentPayloadType = t.payloadTypeH265
		case VideoTrackCodecVP8:
			t.currentPayloadType = t.payloadTypeVP8
		case VideoTrackCodecVP9:
			t.currentPayloadType = t.payloadTypeVP9
		case VideoTrackCodecAV1:
			t.currentPayloadType = t.payloadTypeAV1
		case audioTrackCodecOpus:
			t.currentPayloadType = t.payloadTypeOpus
		}
	}

	packet.PayloadType = t.currentPayloadType

	if _, err := t.writeStream.WriteRTP(&packet.Header, packet.Payload); err != nil {
		t.errorCount += 1

		if t.errorCount%50 == 0 {
			log.Println("WHIPSession.TrackMultiCodec.WriteRTP.Error(", t.errorCount, ")", err)
			return err
		}
	}

	return nil
}
