package session

import (
	"sync/atomic"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/webrtc/v4"
)

type (
	WhepSession struct {
		SessionId            string
		IsWaitingForKeyframe atomic.Bool
		SSEChannel           chan any

		VideoTrack          *codecs.TrackMultiCodec
		VideoTimestamp      uint32
		VideoPacketsWritten uint64
		VideoSequenceNumber uint16
		VideoLayerCurrent   atomic.Value

		AudioTrack          *codecs.TrackMultiCodec
		AudioTimestamp      uint32
		AudioPacketsWritten uint64
		AudioSequenceNumber uint16
		AudioLayerCurrent   atomic.Value

		OnTrackHandler func(stream *WhipSession) func(*webrtc.TrackRemote, *webrtc.RTPReceiver)
	}
)
