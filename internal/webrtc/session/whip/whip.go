package whip

import (
	"log"
	"maps"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/session/whep"
)

// Start a status loop for the whip session.
//
// - Initializes by announcing stream start to potentially awaiting clients
// - Announces layers changes to clients when layers are added or removed from the session
// - Triggers a status update every 5 seconds to send to all listening WHEP sessions
func (whipSession *WhipSession) StartWhipSessionStatusLoop() {
	log.Println("WhipSession.StartWhipSessionStatusLoop")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	whipSession.AnnounceStreamStartToWhepClients()

	for {
		select {

		// Whip session is shutting down
		case <-whipSession.ActiveContext.Done():
			whipSession.handleAnnounceOffline()
			log.Println("WhipSession.StartWhipSessionStatusLoop.Done")
			return

		// Announce new layers available
		case <-whipSession.OnTrackChangeChannel:
			log.Println("WhipSession.AnnounceLayersToWhepClients")
			whipSession.AnnounceLayersToWhepClients()

		// Send status every 5 seconds
		case <-ticker.C:
			whipSession.handleStatus()
		}
	}
}

// Start a routing that takes snapshots of the current whep sessions in the whip session.
func (whipSession *WhipSession) Snapshot() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			whipSession.WhepSessionsLock.RLock()
			snapshot := make(map[string]*whep.WhepSession, len(whipSession.WhepSessions))
			for _, whepSession := range whipSession.WhepSessions {
				if whepSession.IsSessionClosed.Load() == false {
					snapshot[whepSession.SessionId] = whepSession
				}
			}
			whipSession.WhepSessionsLock.RUnlock()

			whipSession.WhepSessionsSnapshot.Store(snapshot)
		case <-whipSession.ActiveContext.Done():
			whipSession.WhepSessionsSnapshot.Store(nil)
			return
		}
	}
}
