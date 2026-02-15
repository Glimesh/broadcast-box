package whep

import (
	"log"
	"time"
)

func (whepSession *WhepSession) handleCalculatedValues() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastBytesReceived := int(0)

	for {
		select {
		case <-whepSession.ActiveContext.Done():
			log.Println("WhepSession.HandleCalculatedValues.Close")
			return
		case <-ticker.C:
			whepSession.VideoLock.RLock()
			videoBytesWritten := whepSession.VideoBytesWritten
			whepSession.VideoLock.RUnlock()

			whepSession.VideoBitrate.Store(uint64(videoBytesWritten - lastBytesReceived))
			lastBytesReceived = videoBytesWritten
		}
	}
}
