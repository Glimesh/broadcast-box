package whep

import "log"

// Handle events for SSE to the WHEP sessions
func (whepSession *WhepSession) handleEvents() {
	for {
		select {
		case <-whepSession.ActiveContext.Done():
			log.Println("WhepSession.HandleEventsLoop.Close")
			return
		case msg, ok := <-whepSession.WhipEventsChannel:
			if !ok {
				log.Println("WhepSession.Event.Whip: Channel closed - exiting")
			} else {
				log.Println("WhepSession.Event.Whip:", msg)
			}
		}
	}
}

// Handles incoming stream packets
func (whepSession *WhepSession) handleStream() {
	for {
		select {
		case <-whepSession.ActiveContext.Done():
			log.Println("WhepSession.HandleStreamLoop.Close")
			return

		case packet := <-whepSession.VideoChannel:
			if packet.Layer == whepSession.VideoLayerCurrent.Load() {
				whepSession.SendVideoPacket(packet)
			}

		case packet := <-whepSession.AudioChannel:
			if packet.Layer == whepSession.AudioLayerCurrent.Load() {
				whepSession.SendAudioPacket(packet)
			}
		}
	}
}
