package webrtc

import (
	"os"
	"strings"

	"github.com/pion/webrtc/v4"
)

func CreatePeerConnection(api *webrtc.API) (*webrtc.PeerConnection, error) {
	config := webrtc.Configuration{}

	if stunServers := os.Getenv("STUN_SERVERS"); stunServers != "" {
		for stunServer := range strings.SplitSeq(stunServers, "|") {
			config.ICEServers = append(config.ICEServers, webrtc.ICEServer{
				URLs: []string{"stun:" + stunServer},
			})
		}
	}

	return api.NewPeerConnection(config)
}
