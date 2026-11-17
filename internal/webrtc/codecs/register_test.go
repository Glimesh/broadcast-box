package codecs

import (
	"strings"
	"testing"

	"github.com/pion/webrtc/v4"
)

func TestRegisterCodecsOffersVP8(t *testing.T) {
	mediaEngine := &webrtc.MediaEngine{}
	RegisterCodecs(mediaEngine)
	peer, err := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine)).NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := peer.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := peer.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		t.Fatal(err)
	}
	offer, err := peer.CreateOffer(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(offer.SDP, "a=rtpmap:97 VP8/90000") {
		t.Fatal("VP8 is missing from video offer")
	}
}
