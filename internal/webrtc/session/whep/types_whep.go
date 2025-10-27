package whep

import (
	"sync"
	"sync/atomic"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type TrackPacket struct {
	Layer        string
	Packet       *rtp.Packet
	TimeDiff     int64
	SequenceDiff int
	Codec        codecs.TrackCodeType
	IsKeyframe   bool
}

type (
	WhepSession struct {
		SessionId            string
		IsWaitingForKeyframe atomic.Bool
		IsSessionClosed      atomic.Bool

		WhipEventsChannel    chan any
		SseEventsChannel     chan any
		SessionClose         sync.Once
		SessionClosedChannel chan struct{}

		PeerConnection *webrtc.PeerConnection

		// Protects VideoTrack, VideoTimestamp, VideoPacketsWritten, VideoSequenceNumber
		VideoLock           sync.RWMutex
		VideoTrack          *codecs.TrackMultiCodec
		VideoTimestamp      uint32
		VideoPacketsWritten uint64
		VideoSequenceNumber uint16
		VideoLayerCurrent   atomic.Value
		VideoChannel        chan TrackPacket

		// Protects AudioTrack, AudioTimestamp, AudioPacketsWritten, AudioSequenceNumber
		AudioLock           sync.RWMutex
		AudioTrack          *codecs.TrackMultiCodec
		AudioTimestamp      uint32
		AudioPacketsWritten uint64
		AudioSequenceNumber uint16
		AudioLayerCurrent   atomic.Value
		AudioChannel        chan TrackPacket
	}
)

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
