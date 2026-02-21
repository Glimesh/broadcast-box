package chat

import (
	"os"
	"strconv"
	"time"
)

const (
	DefaultMaxHistory      = 10000
	DefaultTTL             = 72 * time.Hour
	DefaultCleanupInterval = 1 * time.Hour

	EventTypeMessage = "message"
)

type Message struct {
	ID          string `json:"id"`
	TS          int64  `json:"ts"`
	Text        string `json:"text"`
	DisplayName string `json:"displayName"`
}

type Event struct {
	ID      uint64  `json:"-"`
	Type    string  `json:"type"`
	Message Message `json:"message"`
}

type Session struct {
	ID           string
	StreamKey    string
	LastActivity time.Time
}

type Store interface {
	Connect(streamKey string, now time.Time) string
	GetSession(sessionID string, now time.Time) (*Session, bool)
	TouchSession(sessionID string, now time.Time) bool
	Subscribe(sessionID string, lastEventID uint64, now time.Time) (chan Event, func(), []Event, error)
	SubscribeStream(streamKey string, lastEventID uint64, now time.Time) (chan Event, func(), []Event, error)
	Send(sessionID string, text string, displayName string, now time.Time) error
	SendToStream(streamKey string, text string, displayName string, now time.Time) error
	Cleanup(now time.Time, ttl time.Duration)
}

type Manager struct {
	store           Store
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	stop            chan struct{}
}

func NewManager() *Manager {
	maxHistory := DefaultMaxHistory
	if val := os.Getenv("CHAT_MAX_HISTORY"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			maxHistory = i
		}
	}

	defaultTTL := DefaultTTL
	if val := os.Getenv("CHAT_DEFAULT_TTL"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			defaultTTL = d
		}
	}

	cleanupInterval := DefaultCleanupInterval
	if val := os.Getenv("CHAT_CLEANUP_INTERVAL"); val != "" {
		if d, err := time.ParseDuration(val); err == nil && d > 0 {
			cleanupInterval = d
		}
	}

	m := NewManagerWithStore(NewInMemoryStore(maxHistory), defaultTTL, cleanupInterval)

	return m
}

func NewManagerWithStore(store Store, defaultTTL time.Duration, cleanupInterval time.Duration) *Manager {
	m := &Manager{
		store:           store,
		defaultTTL:      defaultTTL,
		cleanupInterval: cleanupInterval,
		stop:            make(chan struct{}),
	}
	go m.cleanupLoop()
	return m
}

func (m *Manager) Connect(streamKey string) string {
	return m.store.Connect(streamKey, time.Now())
}

func (m *Manager) GetSession(sessionID string) (*Session, bool) {
	return m.store.GetSession(sessionID, time.Now())
}

func (m *Manager) TouchSession(sessionID string) bool {
	return m.store.TouchSession(sessionID, time.Now())
}

func (m *Manager) Subscribe(sessionID string, lastEventID uint64) (chan Event, func(), []Event, error) {
	return m.store.Subscribe(sessionID, lastEventID, time.Now())
}

func (m *Manager) Send(sessionID string, text string, displayName string) error {
	return m.store.Send(sessionID, text, displayName, time.Now())
}

func (m *Manager) SubscribeStream(streamKey string, lastEventID uint64) (chan Event, func(), []Event, error) {
	return m.store.SubscribeStream(streamKey, lastEventID, time.Now())
}

func (m *Manager) SendToStream(streamKey string, text string, displayName string) error {
	return m.store.SendToStream(streamKey, text, displayName, time.Now())
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stop:
			return
		}
	}
}

func (m *Manager) Stop() {
	close(m.stop)
}

func (m *Manager) cleanup() {
	m.store.Cleanup(time.Now(), m.defaultTTL)
}
