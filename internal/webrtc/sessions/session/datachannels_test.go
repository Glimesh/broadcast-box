package session

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataChannelBroadcast(t *testing.T) {
	s := &Session{StreamKey: "stream-1"}

	// Register peers
	senderChannel := &fakeDataChannel{}
	recipientChannel := &fakeDataChannel{}
	failingRecipientChannel := &fakeDataChannel{sendTextError: errors.New("send failed")}
	otherStreamChannel := &fakeDataChannel{}

	sender := s.addDataChannelPeer("sender", senderChannel)
	recipient := s.addDataChannelPeer("recipient", recipientChannel)
	failingRecipient := s.addDataChannelPeer("failing-recipient", failingRecipientChannel)
	(&Session{StreamKey: "stream-2"}).addDataChannelPeer("other-stream", otherStreamChannel)

	// Text broadcasts
	s.broadcastDataChannelFrom(sender, []byte("hello"), true)
	assert.Empty(t, senderChannel.textMessages)
	assert.Empty(t, senderChannel.binaryMessages)
	assert.Equal(t, []string{"hello"}, recipientChannel.textMessages)
	assert.Empty(t, recipientChannel.binaryMessages)
	assert.Empty(t, otherStreamChannel.textMessages)
	assert.Empty(t, otherStreamChannel.binaryMessages)

	// Binary broadcasts
	s.broadcastDataChannelFrom(sender, []byte{0x01, 0x02, 0x03}, false)
	assert.Equal(t, [][]byte{{0x01, 0x02, 0x03}}, recipientChannel.binaryMessages)
	assert.Equal(t, []string{"hello"}, recipientChannel.textMessages)

	// Send failures
	assert.True(t, s.isDataChannelPeerRegistered(failingRecipient))
	s.broadcastDataChannelFrom(failingRecipient, []byte("still active"), true)
	assert.True(t, s.isDataChannelPeerRegistered(failingRecipient))
	assert.Equal(t, []string{"hello", "still active"}, recipientChannel.textMessages)

	// Unregister a peer
	s.removeDataChannelPeer(recipient)
	s.broadcastDataChannelFrom(sender, []byte("after unregister"), true)
	assert.Equal(t, []string{"hello", "still active"}, recipientChannel.textMessages)

	// Replace a peer
	oldChannel := &fakeDataChannel{}
	newChannel := &fakeDataChannel{}
	oldPeer := s.addDataChannelPeer("duplicate", oldChannel)
	newPeer := s.addDataChannelPeer("duplicate", newChannel)

	s.removeDataChannelPeer(oldPeer)
	assert.True(t, s.isDataChannelPeerRegistered(newPeer))
	s.broadcastDataChannelFrom(sender, []byte("replacement"), true)
	assert.Empty(t, oldChannel.textMessages)
	assert.Equal(t, []string{"replacement"}, newChannel.textMessages)
}

func (s *Session) isDataChannelPeerRegistered(peer *dataChannelPeer) bool {
	s.dataChannelPeersLock.RLock()
	defer s.dataChannelPeersLock.RUnlock()
	return s.dataChannelPeers[peer.id] == peer
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
