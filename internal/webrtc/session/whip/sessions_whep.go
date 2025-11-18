package whip

import "log"

// Remove all WHEP sessions from the WHIP session.
func (whipSession *WhipSession) RemoveWhepSessions() {
	log.Println("WhipSession.RemoveWhepSessions:", whipSession.StreamKey)
	whipSession.WhepSessionsLock.Lock()

	for _, whepSession := range whipSession.WhepSessions {
		whepSession.ActiveContextCancel()
		delete(whipSession.WhepSessions, whepSession.SessionId)
	}

	whipSession.WhepSessionsLock.Unlock()
}

// Remove a WHEP session from the WHIP session.
// If the WHIP session no longer has a host, or any WHEP sessions, terminate it.
func (whipSession *WhipSession) RemoveWhepSession(whepSessionId string) {
	log.Println("WhipSession.RemoveWhepSession:", whepSessionId)
	whipSession.WhepSessionsLock.Lock()

	if whepSession, ok := whipSession.WhepSessions[whepSessionId]; ok {
		// Close out Whep session and remove
		whepSession.ActiveContextCancel()
		delete(whipSession.WhepSessions, whepSessionId)
	}

	whipSession.WhepSessionsLock.Unlock()

}
