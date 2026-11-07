package codecs

import (
	"sync"
	"testing"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type trackContext struct {
	webrtc.TrackLocalContext
	writer *trackWriter
}

func (c trackContext) SSRC() webrtc.SSRC                    { return 123 }
func (c trackContext) WriteStream() webrtc.TrackLocalWriter { return c.writer }
func (c trackContext) CodecParameters() []webrtc.RTPCodecParameters {
	return []webrtc.RTPCodecParameters{
		{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, PayloadType: 102},
		{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, PayloadType: 96},
	}
}

type trackWriter struct {
	header rtp.Header
	writes int
}

func (w *trackWriter) WriteRTP(header *rtp.Header, payload []byte) (int, error) {
	w.header = *header
	w.writes++
	return len(payload), nil
}
func (w *trackWriter) Write(b []byte) (int, error) { return len(b), nil }

func TestTrackMultiCodecBinding(t *testing.T) {
	track := CreateTrackMultiCodec("id", "", "stream", webrtc.RTPCodecTypeVideo, VideoTrackCodecH264)
	packet := &rtp.Packet{}
	if written, err := track.WriteRTP(packet, VideoTrackCodecH264); err != nil || written {
		t.Fatal(err)
	}

	writer := &trackWriter{}
	ctx := trackContext{writer: writer}
	if _, err := track.Bind(ctx); err != nil {
		t.Fatal(err)
	}
	if written, err := track.WriteRTP(packet, VideoTrackCodecH264); err != nil || !written {
		t.Fatal(err)
	}
	if writer.writes != 1 || writer.header.SSRC != 123 || writer.header.PayloadType != 102 {
		t.Fatalf("unexpected first write: %+v, writes: %d", writer.header, writer.writes)
	}
	if err := track.Unbind(ctx); err != nil {
		t.Fatal(err)
	}
	if written, err := track.WriteRTP(packet, VideoTrackCodecVP8); err != nil || written {
		t.Fatal(err)
	}
	if writer.writes != 1 {
		t.Fatalf("wrote packet after unbind: %d writes", writer.writes)
	}
	if _, err := track.Bind(ctx); err != nil {
		t.Fatal(err)
	}
	if written, err := track.WriteRTP(packet, VideoTrackCodecVP8); err != nil || !written {
		t.Fatal(err)
	}
	if writer.writes != 2 || writer.header.PayloadType != 96 {
		t.Fatalf("unexpected write after rebind: %+v, writes: %d", writer.header, writer.writes)
	}
}

func TestTrackMultiCodecConcurrentBinding(t *testing.T) {
	track := CreateTrackMultiCodec("id", "", "stream", webrtc.RTPCodecTypeVideo, VideoTrackCodecH264)
	ctx := trackContext{writer: &trackWriter{}}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range 100 {
			_, _ = track.Bind(ctx)
			_ = track.Unbind(ctx)
		}
	}()
	go func() {
		defer wg.Done()
		for range 100 {
			_, _ = track.WriteRTP(&rtp.Packet{}, VideoTrackCodecH264)
			_ = track.Kind()
		}
	}()
	wg.Wait()
}
