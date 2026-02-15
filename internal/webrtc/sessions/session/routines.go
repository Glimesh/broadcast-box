package session

import (
	"log"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whip"
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

// When WHEP is established, send initial messages to client
func (session *Session) handleWhepConnection(whipSession *whip.WhipSession, whepSession *whep.WhepSession) {
	log.Println("Session.WhepSession.Connected:", session.StreamKey)
	whepSession.SseEventsChannel <- session.GetSessionStatsEvent()
	whepSession.SseEventsChannel <- whipSession.GetAvailableLayersEvent()

	<-whepSession.ActiveContext.Done()

	log.Println("Session.WhepSession.Disconnected:", session.StreamKey, " - ", whepSession.SessionId)
	session.removeWhep(whepSession.SessionId)
}

// TODO: Implement correctly
// Handle WHEP Layer changes and trigger keyframe from WHIP
// func (session *Session) handleWhepLayerChange(whipSession *whip.WhipSession, whepSession *whep.WhepSession) {
// 	for {
// 		select {
// 		case <-whipSession.ActiveContext.Done():
// 			return
// 		default:
// 			if whepSession.IsSessionClosed.Load() {
// 				return
// 			} else if session.HasHost.Load() && whepSession.IsWaitingForKeyframe.Load() {
// 				log.Println("WhepSession.PictureLossIndication.IsWaitingForKeyframe")
// 				select {
// 				case whipSession.PacketLossIndicationChannel <- true:
// 				default:
// 					log.Println("WhepSession.PictureLossIndication.Channel: Full channel, skipping")
// 				}
// 			}
// 		}
//
// 		time.Sleep(500 * time.Millisecond)
// 	}
// }

func (session *Session) handleWhepVideoRtcpSender(rtcpSender *webrtc.RTPSender) {
	for {
		rtcpPackets, _, rtcpErr := rtcpSender.ReadRTCP()
		if rtcpErr != nil {
			log.Println("WhepSession.ReadRTCP.Error:", rtcpErr)
			return
		}

		host := session.Host.Load()
		if host == nil {
			continue
		}

		for _, packet := range rtcpPackets {
			if _, isPLI := packet.(*rtcp.PictureLossIndication); isPLI {
				select {
				case host.PacketLossIndicationChannel <- true:
				default:
				}
			}
		}
	}
}

// Handle picture loss indication packages
func (session *Session) handleWhepChannels(whepSession *whep.WhepSession) {
	for {
		select {
		case <-whepSession.ActiveContext.Done():
			return

		case <-whepSession.ConnectionChannel:
			host := session.Host.Load()
			if host == nil {
				continue
			}

			select {
			case host.PacketLossIndicationChannel <- true:
			default:
			}
		}
	}
}

// - Initializes by announcing stream start to potentially awaiting clients
// - Announces layers changes to clients when layers are added or removed from the session
// - Triggers a status update every 5 seconds to send to all listening WHEP sessions
func (session *Session) hostStatusLoop() {
	log.Println("Session.Host.HostStatusLoop")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		host := session.Host.Load()
		if host == nil {
			if session.isEmpty() {
				session.close()
				return
			}

			time.Sleep(5 * time.Second)
			continue
		}

		select {

		case <-host.ActiveContext.Done():
			session.RemoveHost()

			if session.isEmpty() {
				session.close()
			}
			return

		// Send status every 5 seconds
		case <-ticker.C:
			if session.isEmpty() {
				session.close()
			} else if session.Host.Load() != nil {

				status := session.GetSessionStatsEvent()
				session.WhepSessionsLock.RLock()
				for _, whep := range session.WhepSessions {
					whep.SseEventsChannel <- status
				}
				session.WhepSessionsLock.RUnlock()

			}
		}
	}
}

// Start a routing that takes snapshots of the current whep sessions in the whip session.
func (session *Session) Snapshot() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-session.ActiveContext.Done():
			if host := session.Host.Load(); host != nil {
				host.WhepSessionsSnapshot.Store(make(map[string]*whep.WhepSession))
			}
			return
		case <-ticker.C:
			if host := session.Host.Load(); host != nil {
				session.WhepSessionsLock.RLock()
				snapshot := make(map[string]*whep.WhepSession, len(session.WhepSessions))

				for _, whepSession := range session.WhepSessions {
					if !whepSession.IsSessionClosed.Load() {
						snapshot[whepSession.SessionId] = whepSession
					}
				}
				session.WhepSessionsLock.RUnlock()

				host.WhepSessionsSnapshot.Store(snapshot)
			}
		}
	}
}
