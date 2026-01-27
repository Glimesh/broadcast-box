package whep

type WhepSessionStateDto struct {
	Id string `json:"id"`

	AudioLayerCurrent   string `json:"audioLayerCurrent"`
	AudioTimestamp      uint32 `json:"audioTimestamp"`
	AudioPacketsWritten uint64 `json:"audioPacketsWritten"`
	AudioSequenceNumber uint64 `json:"audioSequenceNumber"`

	VideoLayerCurrent   string `json:"videoLayerCurrent"`
	VideoTimestamp      uint32 `json:"videoTimestamp"`
	VideoBitrate        uint64 `json:"videoBitrate"`
	VideoPacketsDropped uint64 `json:"videoPacketsDropped"`
	VideoPacketsWritten uint64 `json:"videoPacketsWritten"`
	VideoSequenceNumber uint64 `json:"videoSequenceNumber"`
}
