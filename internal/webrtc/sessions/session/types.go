package session

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/glimesh/broadcast-box/internal/chat"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whep"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/whip"
)

type Session struct {

	// Protects StreamKey, MOTD, HasHost, IsPublic
	StatusLock sync.RWMutex
	StreamKey  string

	MOTD        string
	HasHost     atomic.Bool
	IsPublic    bool
	StreamStart time.Time

	Host atomic.Pointer[whip.WHIPSession]

	closeOnce sync.Once
	onClose   func()

	// Protects WHEPSessions
	WHEPSessionsLock sync.RWMutex
	WHEPSessions     map[string]*whep.WHEPSession

	ChatManager *chat.Manager
}
