package stream

import (
	"context"
	"sync"
	"sync/atomic"
)

type (
	WhipSession struct {
		StreamKey           string
		SessionId           string
		MOTD                string
		ActiveContext       context.Context
		ActiveContextCancel func()
		PliChan             chan any
		HasHost             atomic.Bool
		IsPublic            bool

		VideoTracks []*VideoTrack
		AudioTracks []*AudioTrack

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
		Track           *TrackMultiCodec
	}
	AudioTrack struct {
		Rid             string
		SessionId       string
		Codec           int
		PacketsReceived atomic.Uint64
		LastRecieved    atomic.Value
		Track           *TrackMultiCodec
	}
)
