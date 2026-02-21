package chat

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type subscriber struct {
	ch chan Event
}

type room struct {
	mu           sync.Mutex
	subscribers  map[string]*subscriber
	history      []Event
	nextEventID  uint64
	lastActivity time.Time
}

type InMemoryStore struct {
	mu         sync.RWMutex
	rooms      map[string]*room
	sessions   map[string]*Session
	maxHistory int
}

func NewInMemoryStore(maxHistory int) *InMemoryStore {
	if maxHistory <= 0 {
		maxHistory = DefaultMaxHistory
	}

	return &InMemoryStore{
		rooms:      make(map[string]*room),
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
	r, ok := s.rooms[session.StreamKey]
	s.mu.Unlock()

	if !ok {
		return nil, nil, nil, fmt.Errorf("room not found")
	}

	return s.subscribeToRoom(r, lastEventID, now)
}

func (s *InMemoryStore) SubscribeStream(streamKey string, lastEventID uint64, now time.Time) (chan Event, func(), []Event, error) {
	s.mu.Lock()
	r := s.getOrCreateRoomLocked(streamKey, now)
	s.mu.Unlock()

	return s.subscribeToRoom(r, lastEventID, now)
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
	r, ok := s.rooms[streamKey]
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("room not found")
	}

	s.sendToRoom(r, text, displayName, now)
	return nil
}

func (s *InMemoryStore) SendToStream(streamKey string, text string, displayName string, now time.Time) error {
	s.mu.Lock()
	r := s.getOrCreateRoomLocked(streamKey, now)
	s.mu.Unlock()

	s.sendToRoom(r, text, displayName, now)
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

	for key, r := range s.rooms {
		r.mu.Lock()
		if len(r.subscribers) == 0 && now.Sub(r.lastActivity) > ttl {
			delete(s.rooms, key)
		}
		r.mu.Unlock()
	}
}

func (s *InMemoryStore) subscribeToRoom(r *room, lastEventID uint64, now time.Time) (chan Event, func(), []Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastActivity = now
	subID := uuid.New().String()
	ch := make(chan Event, 100)
	r.subscribers[subID] = &subscriber{ch: ch}

	var history []Event
	if lastEventID > 0 {
		count := 0
		for _, ev := range r.history {
			if ev.ID > lastEventID {
				count++
			}
		}
		history = make([]Event, 0, count)
		for _, ev := range r.history {
			if ev.ID > lastEventID {
				history = append(history, ev)
			}
		}
	} else {
		history = make([]Event, len(r.history))
		copy(history, r.history)
	}

	cleanup := func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		sub, ok := r.subscribers[subID]
		if !ok {
			return
		}

		delete(r.subscribers, subID)
		close(sub.ch)
	}

	return ch, cleanup, history, nil
}

func (s *InMemoryStore) sendToRoom(r *room, text string, displayName string, now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastActivity = now
	event := Event{
		ID:   r.nextEventID,
		Type: EventTypeMessage,
		Message: Message{
			ID:          uuid.New().String(),
			TS:          now.UnixMilli(),
			Text:        text,
			DisplayName: displayName,
		},
	}
	r.nextEventID++

	if len(r.history) >= s.maxHistory {
		r.history = append(r.history[1:], event)
	} else {
		r.history = append(r.history, event)
	}

	for _, sub := range r.subscribers {
		select {
		case sub.ch <- event:
		default:
		}
	}
}

func (s *InMemoryStore) getOrCreateRoomLocked(streamKey string, now time.Time) *room {
	r, ok := s.rooms[streamKey]
	if ok {
		r.lastActivity = now
		return r
	}

	r = &room{
		subscribers:  make(map[string]*subscriber),
		history:      make([]Event, 0, s.maxHistory),
		nextEventID:  1,
		lastActivity: now,
	}
	s.rooms[streamKey] = r

	return r
}
