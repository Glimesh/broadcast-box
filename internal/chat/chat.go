package chat

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultMaxHistory      = 10000
	DefaultTTL             = 72 * time.Hour
	DefaultCleanupInterval = 1 * time.Hour

	EventTypeMessage   = "message"
	EventTypeConnected = "connected"
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
	for range ticker.C {
		m.cleanup()
	}
}

func (m *Manager) cleanup() {
	m.store.Cleanup(time.Now(), m.defaultTTL)
}

type InMemoryStore struct {
	mu         sync.RWMutex
	rooms      map[string]*Room
	sessions   map[string]*Session
	maxHistory int
}

func NewInMemoryStore(maxHistory int) *InMemoryStore {
	if maxHistory <= 0 {
		maxHistory = DefaultMaxHistory
	}

	return &InMemoryStore{
		rooms:      make(map[string]*Room),
		sessions:   make(map[string]*Session),
		maxHistory: maxHistory,
	}
}

func (s *InMemoryStore) Connect(streamKey string, now time.Time) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := uuid.New().String()
	s.sessions[sessionID] = &Session{
		ID:           sessionID,
		StreamKey:    streamKey,
		LastActivity: now,
	}

	s.getOrCreateRoomLocked(streamKey, now)

	return sessionID
}

func (s *InMemoryStore) GetSession(sessionID string, now time.Time) (*Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, false
	}

	session.LastActivity = now
	copy := *session
	return &copy, true
}

func (s *InMemoryStore) TouchSession(sessionID string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return false
	}

	session.LastActivity = now
	return true
}

func (s *InMemoryStore) Subscribe(sessionID string, lastEventID uint64, now time.Time) (chan Event, func(), []Event, error) {
	s.mu.Lock()
	session, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		return nil, nil, nil, fmt.Errorf("invalid session")
	}

	session.LastActivity = now
	room, ok := s.rooms[session.StreamKey]
	s.mu.Unlock()

	if !ok {
		return nil, nil, nil, fmt.Errorf("room not found")
	}

	return s.subscribeToRoom(room, lastEventID, now)
}

func (s *InMemoryStore) SubscribeStream(streamKey string, lastEventID uint64, now time.Time) (chan Event, func(), []Event, error) {
	s.mu.Lock()
	room := s.getOrCreateRoomLocked(streamKey, now)
	s.mu.Unlock()

	return s.subscribeToRoom(room, lastEventID, now)
}

func (s *InMemoryStore) Send(sessionID string, text string, displayName string, now time.Time) error {
	s.mu.Lock()
	session, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("invalid session")
	}

	session.LastActivity = now
	streamKey := session.StreamKey
	room, ok := s.rooms[streamKey]
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("room not found")
	}

	s.sendToRoom(room, text, displayName, now)
	return nil
}

func (s *InMemoryStore) SendToStream(streamKey string, text string, displayName string, now time.Time) error {
	s.mu.Lock()
	room := s.getOrCreateRoomLocked(streamKey, now)
	s.mu.Unlock()

	s.sendToRoom(room, text, displayName, now)
	return nil
}

func (s *InMemoryStore) Cleanup(now time.Time, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, session := range s.sessions {
		if now.Sub(session.LastActivity) > ttl {
			delete(s.sessions, id)
		}
	}

	for key, room := range s.rooms {
		room.mu.Lock()
		if len(room.subscribers) == 0 && now.Sub(room.lastActivity) > ttl {
			delete(s.rooms, key)
		}
		room.mu.Unlock()
	}
}

func (s *InMemoryStore) subscribeToRoom(room *Room, lastEventID uint64, now time.Time) (chan Event, func(), []Event, error) {
	room.mu.Lock()
	defer room.mu.Unlock()

	room.lastActivity = now
	subID := uuid.New().String()
	ch := make(chan Event, 100)
	ch <- Event{Type: EventTypeConnected}
	room.subscribers[subID] = &subscriber{ch: ch}

	var history []Event
	if lastEventID > 0 {
		// Count matching events first to pre-allocate
		count := 0
		for _, ev := range room.history {
			if ev.ID > lastEventID {
				count++
			}
		}
		history = make([]Event, 0, count)
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

		sub, ok := room.subscribers[subID]
		if !ok {
			return
		}

		delete(room.subscribers, subID)
		close(sub.ch)
	}

	return ch, cleanup, history, nil
}

func (s *InMemoryStore) sendToRoom(room *Room, text string, displayName string, now time.Time) {
	room.mu.Lock()
	defer room.mu.Unlock()

	room.lastActivity = now
	event := Event{
		ID:   room.nextEventID,
		Type: EventTypeMessage,
		Message: Message{
			ID:          uuid.New().String(),
			TS:          now.UnixMilli(),
			Text:        text,
			DisplayName: displayName,
		},
	}
	room.nextEventID++

	if len(room.history) >= s.maxHistory {
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
}

func (s *InMemoryStore) getOrCreateRoomLocked(streamKey string, now time.Time) *Room {
	room, ok := s.rooms[streamKey]
	if ok {
		room.lastActivity = now
		return room
	}

	room = &Room{
		streamKey:    streamKey,
		subscribers:  make(map[string]*subscriber),
		history:      make([]Event, 0, s.maxHistory),
		nextEventID:  1,
		lastActivity: now,
	}
	s.rooms[streamKey] = room

	return room
}
