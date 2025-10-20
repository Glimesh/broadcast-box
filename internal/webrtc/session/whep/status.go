package whep

import (
	"log"
)

func (whepSession *WhepSession) GetWhepSessionStatus() (state WhepSessionState) {
	whepSession.AudioLock.RLock()
	whepSession.VideoLock.RLock()

	currentAudioLayer := whepSession.AudioLayerCurrent.Load().(string)
	currentVideoLayer := whepSession.VideoLayerCurrent.Load().(string)

	state = WhepSessionState{
		Id: whepSession.SessionId,

		AudioLayerCurrent:   currentAudioLayer,
		AudioTimestamp:      whepSession.AudioTimestamp,
		AudioPacketsWritten: whepSession.AudioPacketsWritten,
		AudioSequenceNumber: uint64(whepSession.AudioSequenceNumber),

		VideoLayerCurrent:   currentVideoLayer,
		VideoTimestamp:      whepSession.VideoTimestamp,
		VideoPacketsWritten: whepSession.VideoPacketsWritten,
		VideoSequenceNumber: uint64(whepSession.VideoSequenceNumber),
	}

	whepSession.VideoLock.RUnlock()
	whepSession.AudioLock.RUnlock()

	return
}

// Closes down the WHEP session completely
func (whepSession *WhepSession) Close() {
	// Close WHEP channels
	whepSession.SessionClose.Do(func() {
		log.Println("WhepSession.Close")
		whepSession.IsSessionClosed.Store(true)

		// Close PeerConnection
		err := whepSession.PeerConnection.Close()
		if err != nil {
			log.Println("WhepSession.Close.PeerConnection.Error", err)
		}

		// Notify WHIP about closure
		whepSession.SessionClosedChannel <- struct{}{}

		// Should this fire on close, or be triggered when the Whip session is disposed?
		// whepSession.EventsChannel <- "close"

		// Empty tracks
		whepSession.AudioLock.Lock()
		whepSession.VideoLock.Lock()

		whepSession.AudioTrack = nil
		whepSession.VideoTrack = nil

		whepSession.VideoLock.Unlock()
		whepSession.AudioLock.Unlock()

	})
}
