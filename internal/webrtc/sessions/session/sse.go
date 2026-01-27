package session

import (
	"log"
	"maps"

	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
)

// Get SSE String with status about the current session
func (session *Session) GetSessionStatsEvent() string {

	status, err := utils.ToJsonString(session.GetStreamStatus())
	if err != nil {
		log.Println("GetSessionStatsJsonString Error:", err)
		return ""
	}

	return "event: status\ndata: " + status + "\n\n"
}

// Send out an event to all WHEP sessions to notify that available layers has changed
func (session *Session) AnnounceStreamStartToWhepClients() {
	log.Println("Session.AnnounceStreamStartToWhepClients:", session.StreamKey)

	// Lock, copy session data, then unlock
	session.WhepSessionsLock.RLock()
	whepSessionsCopy := make(map[string]*whep.WhepSession)
	maps.Copy(whepSessionsCopy, session.WhepSessions)
	session.WhepSessionsLock.RUnlock()

	// Generate layer info outside lock
	streamStartMessage := "event: streamStart\ndata:\n"

	// Send to each WHEP session
	for _, whepSession := range whepSessionsCopy {
		if !whepSession.IsSessionClosed.Load() {
			// Announce to frontend client
			select {
			case whepSession.SseEventsChannel <- streamStartMessage:
			default:
				log.Println("WhepSession.AnnounceStreamStartToWhepClients: Channel full, skipping update (SessionId:", whepSession.SessionId, ")")
			}
		}
	}
}
