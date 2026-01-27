package whep

import (
	"log"

	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
)

func (whepSession *WhepSession) GetWhepSessionStatusEvent() string {
	currentSessionStateJson, err := utils.ToJsonString(whepSession.GetWhepSessionStatus())
	if err != nil {
		log.Println("WhepSession.GetWhepSessionStatus Error:", err)
		return ""
	}

	return "event: status\ndata: " + currentSessionStateJson + "\n\n"
}
