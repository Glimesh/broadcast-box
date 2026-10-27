package session

import (
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
	"time"
)

// Status for an individual streaming session
type whipSessionStatus struct {
	StreamKey   string    `json:"streamKey"`
	MOTD        string    `json:"motd"`
	ViewerCount int       `json:"viewers"`
	IsOnline    bool      `json:"isOnline"`
	StreamStart time.Time `json:"streamStart"`
}

// Information for a whip session
type StreamSessionState struct {
	StreamKey   string    `json:"streamKey"`
	Username    string    `json:"username,omitempty"`
	DisplayName string    `json:"displayName,omitempty"`
	AvatarURL   string    `json:"avatarUrl,omitempty"`
	IsPublic    bool      `json:"isPublic"`
	MOTD        string    `json:"motd"`
	StreamStart time.Time `json:"streamStart"`

	AudioTracks []AudioTrackState `json:"audioTracks"`
	VideoTracks []VideoTrackState `json:"videoTracks"`

	Sessions []whep.SessionState `json:"sessions"`
}

type AudioTrackState struct {
	Rid             string `json:"rid"`
	PacketsReceived uint64 `json:"packetsReceived"`
	PacketsDropped  uint64 `json:"packetsDropped"`
}

type VideoTrackState struct {
	Rid             string    `json:"rid"`
	Bitrate         uint64    `json:"bitrate"`
	PacketsReceived uint64    `json:"packetsReceived"`
	PacketsDropped  uint64    `json:"packetsDropped"`
	LastKeyframe    time.Time `json:"lastKeyframe"`
}
