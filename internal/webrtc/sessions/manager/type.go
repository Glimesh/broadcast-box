package manager

import (
	"sync"

	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/session"
	"github.com/pion/webrtc/v4"
)

var (
	SessionsManager *SessionManager

	ApiWhip *webrtc.API
	ApiWhep *webrtc.API
)

type SessionManager struct {
	sessionsLock sync.RWMutex
	sessions     map[string]*session.Session
}
