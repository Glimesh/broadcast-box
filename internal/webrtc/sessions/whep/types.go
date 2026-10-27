package whep

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/glimesh/broadcast-box/internal/chat"
	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/webrtc/v4"
)

type (
	WHEPSession struct {
		SessionID            string
		StreamKey            string
		IsWaitingForKeyframe atomic.Bool
		IsSessionClosed      atomic.Bool

		SessionClose sync.Once
		onClose      func(string)
		pliSender    func()

		PeerConnectionLock sync.RWMutex
		PeerConnection     *webrtc.PeerConnection

		// Protects VideoTrack, VideoTimestamp, VideoPacketsWritten, VideoSequenceNumber,
		// and auto video layer selection state.
		VideoLock               sync.RWMutex
		VideoTrack              *codecs.TrackMultiCodec
		VideoTimestamp          uint32
		VideoBitrate            atomic.Uint64
		VideoBytesWritten       int
		videoBitrateWindowStart time.Time
		videoBitrateWindowBytes int
		VideoPacketsWritten     uint64
		VideoPacketsDropped     atomic.Uint64
		VideoSequenceNumber     uint16
		VideoLayerCurrent       atomic.Value
		videoLayerPriority      int
		videoLayerExplicit      bool

		// Protects AudioTrack, AudioTimestamp, AudioPacketsWritten, AudioSequenceNumber
		AudioLock           sync.RWMutex
		AudioTrack          *codecs.TrackMultiCodec
		AudioTimestamp      uint32
		AudioPacketsWritten uint64
		AudioSequenceNumber uint16
		AudioLayerCurrent   atomic.Value

		ChatManager *chat.Manager
	}
)
