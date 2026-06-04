package datadc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataChannelBroadcast(t *testing.T) {
	m := NewManager()

	// Register peers
	senderChannel := &fakeDataChannel{}
	recipientChannel := &fakeDataChannel{}
	failingRecipientChannel := &fakeDataChannel{sendTextError: errors.New("send failed")}
	otherStreamChannel := &fakeDataChannel{}

	sender := m.register("stream-1", "sender", senderChannel)
	recipient := m.register("stream-1", "recipient", recipientChannel)
	failingRecipient := m.register("stream-1", "failing-recipient", failingRecipientChannel)
	m.register("stream-2", "other-stream", otherStreamChannel)

	// Text broadcasts
	m.broadcastFrom(sender, []byte("hello"), true)
	assert.Empty(t, senderChannel.textMessages)
	assert.Empty(t, senderChannel.binaryMessages)
	assert.Equal(t, []string{"hello"}, recipientChannel.textMessages)
	assert.Empty(t, recipientChannel.binaryMessages)
	assert.Empty(t, otherStreamChannel.textMessages)
	assert.Empty(t, otherStreamChannel.binaryMessages)

	// Binary broadcasts
	m.broadcastFrom(sender, []byte{0x01, 0x02, 0x03}, false)
	assert.Equal(t, [][]byte{{0x01, 0x02, 0x03}}, recipientChannel.binaryMessages)
	assert.Equal(t, []string{"hello"}, recipientChannel.textMessages)

	// Send failures
	assert.True(t, m.isRegistered(failingRecipient))
	m.broadcastFrom(failingRecipient, []byte("still active"), true)
	assert.True(t, m.isRegistered(failingRecipient))
	assert.Equal(t, []string{"hello", "still active"}, recipientChannel.textMessages)

	// Unregister a peer
	m.unregister(recipient)
	m.broadcastFrom(sender, []byte("after unregister"), true)
	assert.Equal(t, []string{"hello", "still active"}, recipientChannel.textMessages)

	// Replace a peer
	oldChannel := &fakeDataChannel{}
	newChannel := &fakeDataChannel{}
	oldPeer := m.register("stream-1", "duplicate", oldChannel)
	newPeer := m.register("stream-1", "duplicate", newChannel)

	m.unregister(oldPeer)
	assert.True(t, m.isRegistered(newPeer))
	m.broadcastFrom(sender, []byte("replacement"), true)
	assert.Empty(t, oldChannel.textMessages)
	assert.Equal(t, []string{"replacement"}, newChannel.textMessages)
}

type fakeDataChannel struct {
	textMessages   []string
	binaryMessages [][]byte
	sendTextError  error
	sendError      error
}

func (f *fakeDataChannel) Send(data []byte) error {
	if f.sendError != nil {
		return f.sendError
	}

	f.binaryMessages = append(f.binaryMessages, append([]byte(nil), data...))
	return nil
}

func (f *fakeDataChannel) SendText(s string) error {
	if f.sendTextError != nil {
		return f.sendTextError
	}

	f.textMessages = append(f.textMessages, s)
	return nil
}
