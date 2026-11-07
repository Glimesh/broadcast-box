package whip

import (
	"testing"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/rtp"
	pionCodecs "github.com/pion/rtp/codecs"
)

func TestVP8KeyframeStart(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		want    bool
	}{
		{"keyframe start", []byte{0x10, 0x00, 0x00, 0x00}, true},
		{"keyframe with picture ID", []byte{0x90, 0x80, 0x01, 0x00}, true},
		{"delta frame start", []byte{0x10, 0x01, 0x00, 0x00}, false},
		{"continuation", []byte{0x00, 0x00}, false},
		{"other partition", []byte{0x11, 0x00}, false},
		{"empty frame", []byte{0x10}, false},
		{"truncated descriptor", []byte{0x90, 0x80}, false},
	}
	depacketizer := &pionCodecs.VP8Packet{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packet := &rtp.Packet{Payload: tt.payload}
			if got := isPacketKeyframe(packet, codecs.VideoTrackCodecVP8, depacketizer); got != tt.want {
				t.Fatalf("isPacketKeyframe() = %v, want %v", got, tt.want)
			}
		})
	}
}
