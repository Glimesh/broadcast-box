package webrtc

import "testing"

func TestAudioOnly(t *testing.T) {
	session := &whepSession{
		videoTrack: nil,
		timestamp:  50000,
	}

	session.sendVideoPacket(nil, "", 0, 0, 0, true)
}
