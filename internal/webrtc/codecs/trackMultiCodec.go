package codecs

import (
	"log"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type TrackMultiCodec struct {
	id       string
	rid      string
	streamId string
	kind     webrtc.RTPCodecType
	codec    int

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

func CreateTrackMultiCodec(id string, rid string, streamId string, kind webrtc.RTPCodecType, codec int) *TrackMultiCodec {
	return &TrackMultiCodec{
		id:       id,
		rid:      rid,
		streamId: streamId,
		kind:     kind,
		codec:    codec,
	}
}

func (track *TrackMultiCodec) Bind(ctx webrtc.TrackLocalContext) (webrtc.RTPCodecParameters, error) {
	track.ssrc = ctx.SSRC()
	track.writeStream = ctx.WriteStream()

	var videoCodecParameters webrtc.RTPCodecParameters
	codecParameters := ctx.CodecParameters()
	for parameters := range codecParameters {
		switch GetAudioTrackCodec(codecParameters[parameters].MimeType) {
		case AudioTrackCodecOpus:
			track.payloadTypeOpus = uint8(codecParameters[parameters].PayloadType)
			track.currentPayloadType = track.payloadTypeOpus
		}

		if track.payloadTypeOpus != 0 {
			log.Println("TrackMultiCodec: Binding AudioTrack Type for", track.streamId, "-", track.currentPayloadType)

			track.kind = webrtc.RTPCodecTypeAudio
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
			track.payloadTypeH264 = uint8(codecParameters[parameters].PayloadType)
			track.currentPayloadType = track.payloadTypeH264
			videoCodecParameters = codecParameters[parameters]

		case VideoTrackCodecH265:
			track.payloadTypeH265 = uint8(codecParameters[parameters].PayloadType)
			track.currentPayloadType = track.payloadTypeH265
			videoCodecParameters = codecParameters[parameters]

		case VideoTrackCodecVP8:
			track.payloadTypeVP8 = uint8(codecParameters[parameters].PayloadType)
			track.currentPayloadType = track.payloadTypeVP8
			videoCodecParameters = codecParameters[parameters]

		case VideoTrackCodecVP9:
			track.payloadTypeVP9 = uint8(codecParameters[parameters].PayloadType)
			track.currentPayloadType = track.payloadTypeVP9
			videoCodecParameters = codecParameters[parameters]

		case VideoTrackCodecAV1:
			track.payloadTypeAV1 = uint8(codecParameters[parameters].PayloadType)
			track.currentPayloadType = track.payloadTypeAV1
			videoCodecParameters = codecParameters[parameters]
		}
	}

	log.Println("TrackMultiCodec: Binding VideoTrack Type for", track.streamId, "-", track.currentPayloadType)
	track.kind = webrtc.RTPCodecTypeVideo
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

func (track *TrackMultiCodec) Unbind(webrtc.TrackLocalContext) error {
	return nil
}

func (track *TrackMultiCodec) WriteRTP(packet *rtp.Packet, codec int) error {
	packet.SSRC = uint32(track.ssrc)

	if codec != track.codec {
		log.Println("TrackMultiCodec: Setting Codec on", track.streamId, "(", track.RID(), ")", "from", track.codec, "to", codec)
		track.codec = codec

		switch track.codec {
		case VideoTrackCodecH264:
			track.currentPayloadType = track.payloadTypeH264
		case VideoTrackCodecH265:
			track.currentPayloadType = track.payloadTypeH265
		case VideoTrackCodecVP8:
			track.currentPayloadType = track.payloadTypeVP8
		case VideoTrackCodecVP9:
			track.currentPayloadType = track.payloadTypeVP9
		case VideoTrackCodecAV1:
			track.currentPayloadType = track.payloadTypeAV1
		case AudioTrackCodecOpus:
			track.currentPayloadType = track.payloadTypeOpus
		}
	}

	packet.PayloadType = track.currentPayloadType

	if _, err := track.writeStream.WriteRTP(&packet.Header, packet.Payload); err != nil {
		log.Println("WriteRTP.Error", err)
		return err
	}

	return nil
}

func (track *TrackMultiCodec) ID() string                { return track.id }
func (track *TrackMultiCodec) RID() string               { return track.rid }
func (track *TrackMultiCodec) StreamID() string          { return track.streamId }
func (track *TrackMultiCodec) Kind() webrtc.RTPCodecType { return track.kind }
