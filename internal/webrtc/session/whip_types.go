package session

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/webrtc/v4"
)

var (
	WhipSessions     map[string]*WhipSession
	WhipSessionsLock sync.Mutex
	ApiWhip          *webrtc.API
	ApiWhep          *webrtc.API
)

type (
	WhipSession struct {
		// Protects StreamKey, SessionId, MOTD, HasHost
		StatusLock sync.RWMutex
		StreamKey  string
		SessionId  string
		MOTD       string
		HasHost    atomic.Bool

		ActiveContext       context.Context
		ActiveContextCancel func()

		PliChan      chan any
		IsPublic     bool
		OnOnlineChan chan bool
		OnTrackChan  chan struct{}
		SSEChan      chan any

		// Protects AudioTrack, VideoTracks
		TracksLock  sync.RWMutex
		VideoTracks []*VideoTrack
		AudioTracks []*AudioTrack

		// Protects WhepSessions
		WhepSessionsLock sync.RWMutex
		WhepSessions     map[string]*WhepSession
	}

	VideoTrack struct {
		Rid             string
		SessionId       string
		Codec           int
		PacketsReceived atomic.Uint64
		LastRecieved    atomic.Value
		LastKeyFrame    atomic.Value
		Track           *codecs.TrackMultiCodec
	}
	AudioTrack struct {
		Rid             string
		SessionId       string
		Codec           int
		PacketsReceived atomic.Uint64
		LastRecieved    atomic.Value
		Track           *codecs.TrackMultiCodec
	}
)
