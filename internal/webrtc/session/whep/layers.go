package whep

import "log"

// Finds the corresponding Whip session to the Whep session id and sets the requested audio layer
func (whepSession *WhepSession) SetAudioLayer(encodingId string) {
	whepSession.AudioLayerCurrent.Store(encodingId)
	whepSession.IsWaitingForKeyframe.Store(true)
	log.Println("Setting Audio Layer completed")
}

// Finds the corresponding Whip session to the Whep session id and sets the requested video layer
func (whepSession *WhepSession) SetVideoLayer(encodingId string) {
	whepSession.VideoLayerCurrent.Store(encodingId)
	whepSession.IsWaitingForKeyframe.Store(true)
	log.Println("Setting Video Layer completed")
}
