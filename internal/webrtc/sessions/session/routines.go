package session

import (
	"log/slog"

	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

func (s *Session) handleWHEPVideoRTCPSender(whepSession *whep.WHEPSession, rtcpSender *webrtc.RTPSender) {
	for {
		rtcpPackets, _, rtcpErr := rtcpSender.ReadRTCP()
		if rtcpErr != nil {
			slog.Error("WHEPSession.ReadRTCP.Error", "err", rtcpErr)
			return
		}

		for _, packet := range rtcpPackets {
			if _, isPLI := packet.(*rtcp.PictureLossIndication); isPLI {
				whepSession.SendPLI()
			}
		}
	}
}
