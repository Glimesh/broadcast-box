package whip

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/webrtc/v4"
)

type (
	WhipSession struct {
		Id                  string
		ContextLock         sync.RWMutex
		ActiveContext       context.Context
		ActiveContextCancel func()
		PeerConnectionLock  sync.RWMutex
		PeerConnection      *webrtc.PeerConnection

		OnTrackChangeChannel chan struct{}
		EventsChannel        chan any

		PacketLossIndicationChannel chan bool

		// Protects AudioTrack, VideoTracks
		TracksLock  sync.RWMutex
		VideoTracks map[string]*VideoTrack
		AudioTracks map[string]*AudioTrack

		//TODO: WhepSessionsSnapshot should only contain information about the current state of the session, not
		// references to chans and other types that cannot be json serialized.
		// Create interface for the purpose and use that with the atomic specifically
		WhepSessionsSnapshot atomic.Value
	}

	VideoTrack struct {
		Rid             string
		SessionId       string
		Priority        int
		Bitrate         atomic.Uint64
		PacketsReceived atomic.Uint64
		PacketsDropped  atomic.Uint64
		LastReceived    atomic.Value
		LastKeyFrame    atomic.Value
		Track           *codecs.TrackMultiCodec
	}
	AudioTrack struct {
		Rid             string
		SessionId       string
		Priority        int
		PacketsReceived atomic.Uint64
		PacketsDropped  atomic.Uint64
		LastReceived    atomic.Value
		Track           *codecs.TrackMultiCodec
	}
)
