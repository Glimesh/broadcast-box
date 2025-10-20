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

		ActiveContext       context.Context
		ActiveContextCancel func()
		PeerConnection      *webrtc.PeerConnection
		PeerConnectionLock  sync.RWMutex

		OnTrackChangeChannel        chan struct{}
		EventsChannel               chan any
		PacketLossIndicationChannel chan any

		// Protects AudioTrack, VideoTracks
		TracksLock  sync.RWMutex
		VideoTracks []*VideoTrack
		AudioTracks []*AudioTrack

		// Protects WhepSessions
		WhepSessionsLock sync.RWMutex
		WhepSessions     map[string]*whep.WhepSession
	}

	VideoTrack struct {
		Rid             string
		SessionId       string
		Codec           int
		Priority        int
		PacketsReceived atomic.Uint64
		LastRecieved    atomic.Value
		LastKeyFrame    atomic.Value
		Track           *codecs.TrackMultiCodec
	}
	AudioTrack struct {
		Rid             string
		SessionId       string
		Codec           int
		Priority        int
		PacketsReceived atomic.Uint64
		LastRecieved    atomic.Value
		Track           *codecs.TrackMultiCodec
	}
)
