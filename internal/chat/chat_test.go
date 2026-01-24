package chat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChat(t *testing.T) {
	m := NewManager()
	streamKey := "test-stream"

	// Test Connect
	sessionID := m.Connect(streamKey)
	assert.NotEmpty(t, sessionID)

	session, ok := m.GetSession(sessionID)
	assert.True(t, ok)
	assert.Equal(t, streamKey, session.StreamKey)

	// Test Subscribe
	ch, cleanup, history, err := m.Subscribe(sessionID, 0)
	assert.NoError(t, err)
	assert.Empty(t, history)
	defer cleanup()

	// Test Send
	err = m.Send(sessionID, "hello", "user1")
	assert.NoError(t, err)

	select {
	case event := <-ch:
		assert.Equal(t, "hello", event.Message.Text)
		assert.Equal(t, "user1", event.Message.DisplayName)
		assert.Equal(t, uint64(1), event.ID)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	// Test History
	_, cleanup2, history2, err := m.Subscribe(sessionID, 0)
	assert.NoError(t, err)
	assert.Len(t, history2, 1)
	assert.Equal(t, "hello", history2[0].Message.Text)
	cleanup2()

	// Test Resume
	ch3, cleanup3, history3, err := m.Subscribe(sessionID, 1)
	assert.NoError(t, err)
	assert.Empty(t, history3)

	err = m.Send(sessionID, "world", "user2")
	assert.NoError(t, err)

	select {
	case event := <-ch3:
		assert.Equal(t, "world", event.Message.Text)
		assert.Equal(t, uint64(2), event.ID)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
	cleanup3()
}
