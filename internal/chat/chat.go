package chat

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	MaxHistory      = 10000
	DefaultTTL      = 72 * time.Hour
	CleanupInterval = 1 * time.Hour
)

type Message struct {
	ID          string `json:"id"`
	TS          int64  `json:"ts"`
	Text        string `json:"text"`
	DisplayName string `json:"displayName"`
}

type Event struct {
	ID      uint64  `json:"-"`
	Message Message `json:"message"`
}

type subscriber struct {
	ch chan Event
}

type Room struct {
	streamKey    string
	mu           sync.Mutex
	subscribers  map[string]*subscriber
	history      []Event
	nextEventID  uint64
	lastActivity time.Time
}

type Session struct {
	ID           string
	StreamKey    string
	LastActivity time.Time
}

type Manager struct {
	mu       sync.RWMutex
	rooms    map[string]*Room
	sessions map[string]*Session
}

func NewManager() *Manager {
	m := &Manager{
		rooms:    make(map[string]*Room),
		sessions: make(map[string]*Session),
	}
	go m.cleanupLoop()
	return m
}

func (m *Manager) Connect(streamKey string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	sessionID := uuid.New().String()
	m.sessions[sessionID] = &Session{
		ID:           sessionID,
		StreamKey:    streamKey,
		LastActivity: now,
	}

	if _, ok := m.rooms[streamKey]; !ok {
		m.rooms[streamKey] = &Room{
			streamKey:    streamKey,
			subscribers:  make(map[string]*subscriber),
			history:      make([]Event, 0, MaxHistory),
			nextEventID:  1,
			lastActivity: now,
		}
	}

	return sessionID
}

func (m *Manager) GetSession(sessionID string) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[sessionID]
	if !ok {
		return nil, false
	}

	s.LastActivity = time.Now()
	copy := *s
	return &copy, true
}

func (m *Manager) TouchSession(sessionID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[sessionID]
	if !ok {
		return false
	}

	s.LastActivity = time.Now()
	return true
}

func (m *Manager) Subscribe(sessionID string, lastEventID uint64) (chan Event, func(), []Event, error) {
	now := time.Now()

	m.mu.Lock()
	session, ok := m.sessions[sessionID]
	if !ok {
		m.mu.Unlock()
		return nil, nil, nil, fmt.Errorf("invalid session")
	}
	session.LastActivity = now
	room, ok := m.rooms[session.StreamKey]
	if !ok {
		m.mu.Unlock()
		return nil, nil, nil, fmt.Errorf("room not found")
	}
	m.mu.Unlock()

	room.mu.Lock()
	defer room.mu.Unlock()

	room.lastActivity = now
	subID := uuid.New().String()
	ch := make(chan Event, 100)
	sub := &subscriber{ch: ch}
	room.subscribers[subID] = sub

	var history []Event
	if lastEventID > 0 {
		for _, ev := range room.history {
			if ev.ID > lastEventID {
				history = append(history, ev)
			}
		}
	} else {
		history = make([]Event, len(room.history))
		copy(history, room.history)
	}

	cleanup := func() {
		room.mu.Lock()
		defer room.mu.Unlock()
		delete(room.subscribers, subID)
		close(ch)
	}

	return ch, cleanup, history, nil
}

func (m *Manager) Send(sessionID string, text string, displayName string) error {
	now := time.Now()

	m.mu.Lock()
	session, ok := m.sessions[sessionID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("invalid session")
	}
	session.LastActivity = now
	room, ok := m.rooms[session.StreamKey]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("room not found")
	}
	m.mu.Unlock()

	room.mu.Lock()
	defer room.mu.Unlock()

	room.lastActivity = now
	event := Event{
		ID: room.nextEventID,
		Message: Message{
			ID:          uuid.New().String(),
			TS:          now.UnixMilli(),
			Text:        text,
			DisplayName: displayName,
		},
	}
	room.nextEventID++

	if len(room.history) >= MaxHistory {
		room.history = append(room.history[1:], event)
	} else {
		room.history = append(room.history, event)
	}

	for _, sub := range room.subscribers {
		select {
		case sub.ch <- event:
		default:
			// Subscriber slow, drop message or handle as needed
		}
	}

	return nil
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(CleanupInterval)
	for range ticker.C {
		m.cleanup()
	}
}

func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, s := range m.sessions {
		if now.Sub(s.LastActivity) > DefaultTTL {
			delete(m.sessions, id)
		}
	}

	for key, r := range m.rooms {
		r.mu.Lock()
		if len(r.subscribers) == 0 && now.Sub(r.lastActivity) > DefaultTTL {
			delete(m.rooms, key)
		}
		r.mu.Unlock()
	}
}
