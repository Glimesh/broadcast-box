package status

import "time"

type StreamState struct {
	StreamKey string `json:"streamKey"`
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
