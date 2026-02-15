package session

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whip"
)

type Session struct {

	// Protects StreamKey, SessionId, MOTD, HasHost, IsPublic
	StatusLock sync.RWMutex
	StreamKey  string

	SessionId   string
	MOTD        string
	HasHost     atomic.Bool
	IsPublic    bool
	StreamStart time.Time

	Host atomic.Pointer[whip.WhipSession]

	// Context
	ActiveContext       context.Context
	ActiveContextCancel func()

	// Protects WhepSessions
	WhepSessionsLock sync.RWMutex
	WhepSessions     map[string]*whep.WhepSession
}
