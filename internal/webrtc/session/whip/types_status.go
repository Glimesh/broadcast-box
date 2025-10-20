package whip

import "time"

// Status for an individual streaming session
type WhipSessionStatus struct {
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

	Sessions []WhepSessionState `json:"sessions"`
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

type WhepSessionState struct {
	Id string `json:"id"`

	AudioLayerCurrent   string `json:"audioLayerCurrent"`
	AudioTimestamp      uint32 `json:"audioTimestamp"`
	AudioPacketsWritten uint64 `json:"audioPacketsWritten"`
	AudioSequenceNumber uint64 `json:"audioSequenceNumber"`

	VideoLayerCurrent   string `json:"videoLayerCurrent"`
	VideoTimestamp      uint32 `json:"videoTimestamp"`
	VideoPacketsWritten uint64 `json:"videoPacketsWritten"`
	VideoSequenceNumber uint64 `json:"videoSequenceNumber"`
}
