package peerconnection

import (
	"os"
	"slices"
	"strings"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
	"github.com/pion/webrtc/v4"
)

func CreateWhepPeerConnection() (*webrtc.PeerConnection, error) {
	return session.ApiWhep.NewPeerConnection(GetPeerConfig())
}

func CreateWhipPeerConnection() (*webrtc.PeerConnection, error) {
	return session.ApiWhip.NewPeerConnection(GetPeerConfig())
}

func GetPeerConfig() webrtc.Configuration {
	config := webrtc.Configuration{}
	if stunServers := os.Getenv("STUN_SERVERS_INTERNAL"); stunServers != "" {
		for stunServer := range strings.SplitSeq(stunServers, "|") {
			config.ICEServers = append(config.ICEServers, webrtc.ICEServer{
				URLs: []string{"stun:" + stunServer},
			})
		}
	} else if stunServers := os.Getenv("STUN_SERVERS"); stunServers != "" {
		for stunServer := range strings.SplitSeq(stunServers, "|") {
			config.ICEServers = append(config.ICEServers, webrtc.ICEServer{
				URLs: []string{"stun:" + stunServer},
			})
		}
	}

	username, credential := authorization.GetTURNCredentials()

	if turnServers := os.Getenv("TURN_SERVERS"); turnServers != "" {
		for turnServer := range strings.SplitSeq(turnServers, "|") {
			config.ICEServers = append(config.ICEServers, webrtc.ICEServer{
				URLs:       []string{"turn:" + turnServer},
				Username:   username,
				Credential: credential,
			})
		}
	} else if turnServers := os.Getenv("TURN_SERVERS_INTERNAL"); turnServers != "" {
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

func disconnected(isWhip bool, streamKey string, streamId string) {
	session.WhipSessionsLock.Lock()
	defer session.WhipSessionsLock.Unlock()

	stream, ok := session.WhipSessions[streamKey]
	if !ok {
		return
	}

	stream.WhepSessionsLock.Lock()
	defer stream.WhepSessionsLock.Unlock()

	if !isWhip {
		delete(stream.WhepSessions, streamId)
	} else {
		stream.AudioTracks = slices.DeleteFunc(stream.AudioTracks, func(track *session.AudioTrack) bool {
			return track.SessionId == stream.SessionId
		})

		stream.VideoTracks = slices.DeleteFunc(stream.VideoTracks, func(track *session.VideoTrack) bool {
			return track.SessionId == stream.SessionId
		})

		if stream.SessionId != streamId {
			return
		}

		stream.HasHost.Store(false)
		stream.OnTrackChan <- struct{}{}
	}

	if len(stream.WhepSessions) != 0 || stream.HasHost.Load() {
		return
	}

	stream.ActiveContextCancel()
	delete(session.WhipSessions, streamKey)
}
