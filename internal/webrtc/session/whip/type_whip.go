package whip

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/session/whep"
	"github.com/pion/webrtc/v4"
)

type (
	simulcastLayerResponse struct {
		EncodingId string `json:"encodingId"`
	}
)

type (
	WhipSession struct {
		// Protects StreamKey, SessionId, MOTD, HasHost, IsPublic
		StatusLock sync.RWMutex
		StreamKey  string
		SessionId  string
		MOTD       string
		HasHost    atomic.Bool
		IsPublic   bool

		ContextLock         sync.RWMutex
		ActiveContext       context.Context
		ActiveContextCancel func()
		PeerConnection      *webrtc.PeerConnection
		PeerConnectionLock  sync.RWMutex

		OnTrackChangeChannel chan struct{}
		EventsChannel        chan any

		// TODO: Moving this to the individual video track might be better
		PacketLossIndicationChannel chan bool

		// Protects AudioTrack, VideoTracks
		TracksLock  sync.RWMutex
		VideoTracks map[string]*VideoTrack
		AudioTracks map[string]*AudioTrack

		// Protects WhepSessions
		WhepSessionsLock sync.RWMutex
		WhepSessions     map[string]*whep.WhepSession

		//TODO: WhepSessionsSnapshot should only contain information about the current state of the session, not
		// references to chans and other types that cannot be json serialized.
		// Create interface for the purpose and use that with the atomic specifically
		WhepSessionsSnapshot atomic.Value
	}

	VideoTrack struct {
		Rid             string
		SessionId       string
		Priority        int
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
