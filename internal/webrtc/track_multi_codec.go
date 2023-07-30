package webrtc

import (
	"strings"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type trackMultiCodec struct {
	ssrc        webrtc.SSRC
	writeStream webrtc.TrackLocalWriter

	h264PayloadType, av1PayloadType uint8

	id, rid, streamID string
}

func (t *trackMultiCodec) Bind(ctx webrtc.TrackLocalContext) (webrtc.RTPCodecParameters, error) {
	t.ssrc = ctx.SSRC()
	t.writeStream = ctx.WriteStream()

	codecs := ctx.CodecParameters()
	for i := range codecs {
		if t.av1PayloadType == 0 && strings.Contains(
			strings.ToLower(webrtc.MimeTypeAV1),
			strings.ToLower(codecs[i].MimeType),
		) {
			t.av1PayloadType = uint8(codecs[i].PayloadType)
		}

		if t.h264PayloadType == 0 && strings.Contains(
			strings.ToLower(webrtc.MimeTypeH264),
			strings.ToLower(codecs[i].MimeType),
		) {
			t.h264PayloadType = uint8(codecs[i].PayloadType)
		}

	}

	return webrtc.RTPCodecParameters{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}}, nil
}

func (t *trackMultiCodec) Unbind(webrtc.TrackLocalContext) error {
	return nil
}

func (t *trackMultiCodec) WriteRTP(p *rtp.Packet, isAV1 bool) error {
	p.Header.SSRC = uint32(t.ssrc)

	if isAV1 {
		p.Header.PayloadType = t.av1PayloadType
	} else {
		p.Header.PayloadType = t.h264PayloadType
	}

	_, err := t.writeStream.WriteRTP(&p.Header, p.Payload)
	return err
}

func (t *trackMultiCodec) ID() string       { return t.id }
func (t *trackMultiCodec) RID() string      { return t.rid }
func (t *trackMultiCodec) StreamID() string { return t.streamID }
func (t *trackMultiCodec) Kind() webrtc.RTPCodecType {
	return webrtc.RTPCodecTypeVideo
}
