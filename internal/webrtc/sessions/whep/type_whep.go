package whep

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/webrtc/v4"
)

type (
	WhepSession struct {
		SessionId            string
		IsWaitingForKeyframe atomic.Bool
		IsSessionClosed      atomic.Bool

		WhipEventsChannel   chan any
		SseEventsChannel    chan any
		ConnectionChannel   chan any
		SessionClose        sync.Once
		ActiveContext       context.Context
		ActiveContextCancel func()

		PeerConnection *webrtc.PeerConnection

		// Protects VideoTrack, VideoTimestamp, VideoPacketsWritten, VideoSequenceNumber
		VideoLock           sync.RWMutex
		VideoTrack          *codecs.TrackMultiCodec
		VideoTimestamp      uint32
		VideoBitrate        atomic.Uint64
		VideoBytesWritten   int
		VideoPacketsWritten uint64
		VideoPacketsDropped atomic.Uint64
		VideoSequenceNumber uint16
		VideoLayerCurrent   atomic.Value
		VideoChannel        chan codecs.TrackPacket

		// Protects AudioTrack, AudioTimestamp, AudioPacketsWritten, AudioSequenceNumber
		AudioLock           sync.RWMutex
		AudioTrack          *codecs.TrackMultiCodec
		AudioTimestamp      uint32
		AudioPacketsWritten uint64
		AudioSequenceNumber uint16
		AudioLayerCurrent   atomic.Value
		AudioChannel        chan codecs.TrackPacket
	}
)
