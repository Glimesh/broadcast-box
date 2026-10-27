package session

import (
	"log"

	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

//TODO: Might not neccessary
// Triggered when a host is disconnected
// func (session *Session) handleHostDisconnect() {
// 	log.Println("Session.Host.Disconnected", session.StreamKey)
//
// 	// WHIP host offline
// 	if session.Host != nil {
// 		session.Host.RemovePeerConnection()
// 		session.Host.RemoveTracks()
// 	}
// 	session.handleAnnounceOffline()
//
// }

func (s *Session) handleWHEPVideoRTCPSender(whepSession *whep.WHEPSession, rtcpSender *webrtc.RTPSender) {
	for {
		rtcpPackets, _, rtcpErr := rtcpSender.ReadRTCP()
		if rtcpErr != nil {
			log.Println("WHEPSession.ReadRTCP.Error:", rtcpErr)
			return
		}

		for _, packet := range rtcpPackets {
			if _, isPLI := packet.(*rtcp.PictureLossIndication); isPLI {
				whepSession.SendPLI()
			}
		}
	}
}
