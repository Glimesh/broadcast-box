package session

import (
	"log"

	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
)

// Get SSE String with status about the current session
func (s *Session) GetSessionStatsEvent() string {

	status, err := utils.ToJSONString(s.GetStreamStatus())
	if err != nil {
		log.Println("GetSessionStatsJsonString Error:", err)
		return ""
	}

	return "event: status\ndata: " + status + "\n\n"
}
