package whep

import (
	"log"
)

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
				return
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
			whepSession.Close()
			return

		case packet, ok := <-whepSession.VideoChannel:
			if !ok {
				log.Println("WhepSession.HandleStream.VideoChannel.Error")
				return
			}

			whepSession.SendVideoPacket(packet)

		case packet, ok := <-whepSession.AudioChannel:
			if !ok {
				log.Println("WhepSession.HandleStream.VideoChannel.Error")
				return
			}

			whepSession.SendAudioPacket(packet)
		}
	}
}
