package webrtc

import (
	"github.com/pion/rtp"
)

const (
	naluTypeBitmask = 0x1F

	idrNALUType = 5
	spsNALUType = 7
	ppsNALUType = 8
)

func isKeyframe(pkt *rtp.Packet, codec videoTrackCodec, depacketizer rtp.Depacketizer) bool {
	if codec == videoTrackCodecH264 {
		nalu, err := depacketizer.Unmarshal(pkt.Payload)
		if err != nil || len(nalu) < 6 {
			return false
		}

		firstNaluType := nalu[4] & naluTypeBitmask
		return firstNaluType == idrNALUType || firstNaluType == spsNALUType || firstNaluType == ppsNALUType
	}
	return false
}
