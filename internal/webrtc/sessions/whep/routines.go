package whep

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/glimesh/broadcast-box/internal/environment"
)

func (whepSession *WhepSession) handleCalculatedValues() {
	ticker := time.NewTicker(1 * time.Second)

	lastBytesReceived := int(0)

	for {
		select {
		case <-whepSession.ActiveContext.Done():
			log.Println("WhepSession.HandleCalculatedValues.Close")
			return
		case <-ticker.C:
			whepSession.VideoBitrate.Store(uint64(whepSession.VideoBytesWritten - lastBytesReceived))
			lastBytesReceived = whepSession.VideoBytesWritten
		}
	}
}

func (whepSession *WhepSession) handleVideoChannel() {
	experimentalWhepPacketDeepCloneToChannel := strings.EqualFold(os.Getenv(environment.WHEP_EXPERIMENTAL_DEEPCOPY_PACKETS_TO_CHANNEL), "true")

	if !experimentalWhepPacketDeepCloneToChannel {
		return
	}

	for {
		select {
		case <-whepSession.ActiveContext.Done():
			log.Println("WhepSession.HandleVideoChannel.Close")
			return
		case packet, ok := <-whepSession.VideoChannel:
			if !ok {
				log.Println("WhepSession.HandleCalculatedValues.PacketError")
				return
			}

			whepSession.SendVideoPacket(packet)
		}
	}
}
