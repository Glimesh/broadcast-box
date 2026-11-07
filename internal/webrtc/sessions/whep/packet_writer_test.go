package whep

import (
	"errors"
	"testing"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type videoTrackContext struct {
	webrtc.TrackLocalContext
	writer *videoTrackWriter
}

func (c videoTrackContext) SSRC() webrtc.SSRC                    { return 123 }
func (c videoTrackContext) WriteStream() webrtc.TrackLocalWriter { return c.writer }
func (c videoTrackContext) CodecParameters() []webrtc.RTPCodecParameters {
	return []webrtc.RTPCodecParameters{
		{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, PayloadType: 102},
	}
}

type videoTrackWriter struct {
	writes  int
	err     error
	started chan struct{}
	release chan struct{}
}

func (w *videoTrackWriter) WriteRTP(_ *rtp.Header, payload []byte) (int, error) {
	if w.started != nil {
		close(w.started)
		<-w.release
	}
	w.writes++
	return len(payload), w.err
}
func (w *videoTrackWriter) Write(b []byte) (int, error) { return len(b), nil }

func TestLayerSwitchKeepsWaitingForNewKeyframe(t *testing.T) {
	track := codecs.CreateTrackMultiCodec("id", "", "stream", webrtc.RTPCodecTypeVideo, codecs.VideoTrackCodecH264)
	writer := &videoTrackWriter{started: make(chan struct{}), release: make(chan struct{})}
	if _, err := track.Bind(videoTrackContext{writer: writer}); err != nil {
		t.Fatal(err)
	}
	session := CreateNewWHEP("id", "stream", nil, track, nil, func() {})
	session.SetVideoLayer("old")
	packet := codecs.TrackPacket{Layer: "old", Packet: &rtp.Packet{}, Codec: codecs.VideoTrackCodecH264, IsKeyframe: true}
	done := make(chan struct{})
	go func() {
		session.SendVideoPacket(packet)
		close(done)
	}()
	<-writer.started
	switched := make(chan struct{})
	go func() {
		session.SetVideoLayer("new")
		close(switched)
	}()
	close(writer.release)
	<-done
	<-switched

	// A packet selected before the switch must not clear the new request either.
	session.SendVideoPacket(packet)
	if !session.IsWaitingForKeyframe.Load() || writer.writes != 1 {
		t.Fatalf("waiting = %v, writes = %d", session.IsWaitingForKeyframe.Load(), writer.writes)
	}
}

func TestSendVideoPacketWaitsAfterFailedKeyframe(t *testing.T) {
	track := codecs.CreateTrackMultiCodec("id", "", "stream", webrtc.RTPCodecTypeVideo, codecs.VideoTrackCodecH264)
	writer := &videoTrackWriter{err: errors.New("write failed")}
	if _, err := track.Bind(videoTrackContext{writer: writer}); err != nil {
		t.Fatal(err)
	}
	session := CreateNewWHEP("id", "stream", nil, track, nil, func() {})
	packet := codecs.TrackPacket{Packet: &rtp.Packet{}, Codec: codecs.VideoTrackCodecH264, IsKeyframe: true}
	session.SendVideoPacket(packet)
	if !session.IsWaitingForKeyframe.Load() {
		t.Fatal("failed keyframe cleared waiting state")
	}

	writer.err = nil
	session.SendVideoPacket(packet)
	if session.IsWaitingForKeyframe.Load() {
		t.Fatal("successful keyframe did not clear waiting state")
	}
}

func TestSendVideoPacketWaitsForBoundKeyframe(t *testing.T) {
	track := codecs.CreateTrackMultiCodec("id", "", "stream", webrtc.RTPCodecTypeVideo, codecs.VideoTrackCodecH264)
	plis := 0
	session := CreateNewWHEP("id", "stream", nil, track, nil, func() { plis++ })
	packet := codecs.TrackPacket{Packet: &rtp.Packet{}, Codec: codecs.VideoTrackCodecH264, IsKeyframe: true}
	session.SendVideoPacket(packet)
	if !session.IsWaitingForKeyframe.Load() {
		t.Fatal("unbound keyframe cleared waiting state")
	}
	writer := &videoTrackWriter{}
	ctx := videoTrackContext{writer: writer}
	if _, err := track.Bind(ctx); err != nil {
		t.Fatal(err)
	}
	packet.IsKeyframe = false
	session.SendVideoPacket(packet)
	if plis != 1 || writer.writes != 0 {
		t.Fatalf("non-keyframe: PLIs = %d, writes = %d", plis, writer.writes)
	}
	packet.IsKeyframe = true
	session.SendVideoPacket(packet)
	if writer.writes != 1 || session.IsWaitingForKeyframe.Load() {
		t.Fatalf("keyframe: writes = %d, waiting = %v", writer.writes, session.IsWaitingForKeyframe.Load())
	}
	if err := track.Unbind(ctx); err != nil {
		t.Fatal(err)
	}
	packet.IsKeyframe = false
	session.SendVideoPacket(packet)
	if !session.IsWaitingForKeyframe.Load() {
		t.Fatal("unbound non-keyframe did not restore waiting state")
	}
	if session.VideoPacketsDropped.Load() != 0 {
		t.Fatal("unbound packets counted as write errors")
	}
}
