package peerconnection

import (
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/pion/webrtc/v4"
)

func GetPeerConnectionConfig() webrtc.Configuration {
	config := webrtc.Configuration{}
	if stunServers := os.Getenv(environment.STUN_SERVERS_INTERNAL); stunServers != "" {
		for stunServer := range strings.SplitSeq(stunServers, "|") {
			config.ICEServers = append(config.ICEServers, webrtc.ICEServer{
				URLs: []string{"stun:" + stunServer},
			})
		}
	} else if stunServers := os.Getenv(environment.STUN_SERVERS); stunServers != "" {
		for stunServer := range strings.SplitSeq(stunServers, "|") {
			config.ICEServers = append(config.ICEServers, webrtc.ICEServer{
				URLs: []string{"stun:" + stunServer},
			})
		}
	}

	username, credential := authorization.GetTURNCredentials()

	if turnServers := os.Getenv(environment.TURN_SERVERS_INTERNAL); turnServers != "" {
		for turnServer := range strings.SplitSeq(turnServers, "|") {
			config.ICEServers = append(config.ICEServers, webrtc.ICEServer{
				URLs:       []string{"turn:" + turnServer},
				Username:   username,
				Credential: credential,
			})
		}
	} else if turnServers := os.Getenv(environment.TURN_SERVERS); turnServers != "" {
		for turnServer := range strings.SplitSeq(turnServers, "|") {
			config.ICEServers = append(config.ICEServers, webrtc.ICEServer{
				URLs:       []string{"turn:" + turnServer},
				Username:   username,
				Credential: credential,
			})
		}
	}

	return config
}
