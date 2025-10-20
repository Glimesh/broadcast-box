package whip

import (
	"log"
	"strings"

	"github.com/pion/sdp/v3"
)

// Helper function for getting the simulcast order and using as priority for consumers
// This example will order from left to right with highest to lowest priority
// a=simulcast:send High,Mid,Low
func (whipSession *WhipSession) getPrioritizedStreamingLayer(layer string, sdpDescription string) int {
	var sessionDescription sdp.SessionDescription
	err := sessionDescription.Unmarshal([]byte(sdpDescription))
	if err != nil {
		log.Println("Track.getPrioritizedStreamingLayer Error: (Layer "+layer+")", err)
		return 100
	}

	var priority = 1
	for _, description := range sessionDescription.MediaDescriptions {
		for _, attribute := range description.Attributes {
			if attribute.Key == "simulcast" && strings.HasPrefix(attribute.Value, "send ") {
				layers := strings.TrimPrefix(attribute.Value, "send")
				for simulcastLayer := range strings.SplitSeq(strings.TrimSpace(layers), ";") {
					if simulcastLayer != "" && strings.EqualFold(simulcastLayer, layer) {
						return priority
					} else {
						priority++
					}
				}
			}
		}
	}

	return 100
}
