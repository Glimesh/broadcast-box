package webrtc

import (
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type trackMultiCodec struct {
	ssrc        webrtc.SSRC
	writeStream webrtc.TrackLocalWriter

	payloadTypeH264, payloadTypeVP8, payloadTypeVP9, payloadTypeAV1 uint8

	id, rid, streamID string
}

func (t *trackMultiCodec) Bind(ctx webrtc.TrackLocalContext) (webrtc.RTPCodecParameters, error) {
	t.ssrc = ctx.SSRC()
	t.writeStream = ctx.WriteStream()

	codecs := ctx.CodecParameters()
	for i := range codecs {
		switch getVideoTrackCodec(codecs[i].MimeType) {
		case videoTrackCodecH264:
			t.payloadTypeH264 = uint8(codecs[i].PayloadType)
		case videoTrackCodecVP8:
			t.payloadTypeVP8 = uint8(codecs[i].PayloadType)
		case videoTrackCodecVP9:
			t.payloadTypeVP9 = uint8(codecs[i].PayloadType)
		case videoTrackCodecAV1:
			t.payloadTypeAV1 = uint8(codecs[i].PayloadType)
		}
	}

	return webrtc.RTPCodecParameters{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}}, nil
}

func (t *trackMultiCodec) Unbind(webrtc.TrackLocalContext) error {
	return nil
}

func (t *trackMultiCodec) WriteRTP(p *rtp.Packet, codec videoTrackCodec) error {
	p.Header.SSRC = uint32(t.ssrc)

	switch codec {
	case videoTrackCodecH264:
		p.Header.PayloadType = t.payloadTypeH264
	case videoTrackCodecVP8:
		p.Header.PayloadType = t.payloadTypeVP8
	case videoTrackCodecVP9:
		p.Header.PayloadType = t.payloadTypeVP9
	case videoTrackCodecAV1:
		p.Header.PayloadType = t.payloadTypeAV1
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
