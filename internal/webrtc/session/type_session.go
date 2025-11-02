package session

import (
	"sync"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/session/whep"
	"github.com/glimesh/broadcast-box/internal/webrtc/session/whip"
	"github.com/pion/webrtc/v4"
)

var (
	SessionManager *WhipSessionManager

	ApiWhip *webrtc.API
	ApiWhep *webrtc.API
)

type WhipSessionManager struct {
	// Protects WhipSessions
	whipSessionsLock sync.RWMutex
	whipSessions     map[string]*whip.WhipSession
}

// Status for an individual streaming session
type StreamStatus struct {
	StreamKey   string `json:"streamKey"`
	MOTD        string `json:"motd"`
	ViewerCount int    `json:"viewers"`
	IsOnline    bool   `json:"isOnline"`
}

// Information for a whip session
type StreamSession struct {
	StreamKey string `json:"streamKey"`
	IsPublic  bool   `json:"isPublic"`
	MOTD      string `json:"motd"`

	AudioTracks []AudioTrackState `json:"audioTracks"`
	VideoTracks []VideoTrackState `json:"videoTracks"`

	Sessions []whep.WhepSessionState `json:"sessions"`
}

type AudioTrackState struct {
	Rid             string `json:"rid"`
	PacketsReceived uint64 `json:"packetsReceived"`
}

type VideoTrackState struct {
	Rid             string    `json:"rid"`
	PacketsReceived uint64    `json:"packetsReceived"`
	LastKeyframe    time.Time `json:"lastKeyframe"`
}
