package stream

import (
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	"log"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
)

type TrackMultiCodec struct {
	id       string
	rid      string
	streamId string
	kind     webrtc.RTPCodecType

	ssrc        webrtc.SSRC
	writeStream webrtc.TrackLocalWriter

	payloadTypeH264 uint8
	payloadTypeH265 uint8
	payloadTypeVP8  uint8
	payloadTypeVP9  uint8
	payloadTypeAV1  uint8
	payloadTypeOpus uint8
}

func NewTrackMultiCodec(id string, rid string, streamId string, kind webrtc.RTPCodecType) *TrackMultiCodec {
	return &TrackMultiCodec{
		id:       id,
		rid:      rid,
		streamId: streamId,
		kind:     kind,
	}
}

func (track *TrackMultiCodec) Bind(ctx webrtc.TrackLocalContext) (webrtc.RTPCodecParameters, error) {
	track.ssrc = ctx.SSRC()
	track.writeStream = ctx.WriteStream()

	var videoCodecParameters webrtc.RTPCodecParameters
	codecParameters := ctx.CodecParameters()
	for parameters := range codecParameters {
		switch codecs.GetAudioTrackCodec(codecParameters[parameters].MimeType) {
		case codecs.AudioTrackCodecOpus:
			track.payloadTypeOpus = uint8(codecParameters[parameters].PayloadType)
		}

		if track.payloadTypeOpus != 0 {

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

		switch codecs.GetVideoTrackCodec(codecParameters[parameters].MimeType) {
		case codecs.VideoTrackCodecH264:
			track.payloadTypeH264 = uint8(codecParameters[parameters].PayloadType)
			videoCodecParameters = codecParameters[parameters]

		case codecs.VideoTrackCodecH265:
			track.payloadTypeH265 = uint8(codecParameters[parameters].PayloadType)
			videoCodecParameters = codecParameters[parameters]

		case codecs.VideoTrackCodecVP8:
			track.payloadTypeVP8 = uint8(codecParameters[parameters].PayloadType)
			videoCodecParameters = codecParameters[parameters]

		case codecs.VideoTrackCodecVP9:
			track.payloadTypeVP9 = uint8(codecParameters[parameters].PayloadType)
			videoCodecParameters = codecParameters[parameters]

		case codecs.VideoTrackCodecAV1:
			track.payloadTypeAV1 = uint8(codecParameters[parameters].PayloadType)
			videoCodecParameters = codecParameters[parameters]

		}
	}

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

	// TODO: Find a better way to negotiate the codec to avoid
	// managing the packet for every write.
	switch codec {
	case codecs.VideoTrackCodecH264:
		packet.PayloadType = track.payloadTypeH264
	case codecs.VideoTrackCodecH265:
		packet.PayloadType = track.payloadTypeH265
	case codecs.VideoTrackCodecVP8:
		packet.PayloadType = track.payloadTypeVP8
	case codecs.VideoTrackCodecVP9:
		packet.PayloadType = track.payloadTypeVP9
	case codecs.VideoTrackCodecAV1:
		packet.PayloadType = track.payloadTypeAV1
	case codecs.AudioTrackCodecOpus:
		packet.PayloadType = track.payloadTypeOpus
	}

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
