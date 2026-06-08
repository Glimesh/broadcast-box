package session

import (
	"errors"
	"testing"

	"github.com/glimesh/broadcast-box/internal/webrtc/datadc"
	"github.com/stretchr/testify/assert"
)

func TestDataChannelBroadcast(t *testing.T) {
	s := &Session{
		StreamKey:        "stream-1",
		DataChannelPeers: map[string]*datadc.Peer{},
	}

	// Register peers
	senderChannel := &fakeDataChannel{}
	recipientChannel := &fakeDataChannel{}
	failingRecipientChannel := &fakeDataChannel{sendTextError: errors.New("send failed")}
	otherStreamChannel := &fakeDataChannel{}

	sender := s.AddDataChannelPeer("sender", senderChannel)
	recipient := s.AddDataChannelPeer("recipient", recipientChannel)
	failingRecipient := s.AddDataChannelPeer("failing-recipient", failingRecipientChannel)
	(&Session{
		StreamKey:        "stream-2",
		DataChannelPeers: map[string]*datadc.Peer{},
	}).AddDataChannelPeer("other-stream", otherStreamChannel)

	// Text broadcasts
	s.BroadcastDataChannelFrom(sender, []byte("hello"), true)
	assert.Empty(t, senderChannel.textMessages)
	assert.Empty(t, senderChannel.binaryMessages)
	assert.Equal(t, []string{"hello"}, recipientChannel.textMessages)
	assert.Empty(t, recipientChannel.binaryMessages)
	assert.Empty(t, otherStreamChannel.textMessages)
	assert.Empty(t, otherStreamChannel.binaryMessages)

	// Binary broadcasts
	s.BroadcastDataChannelFrom(sender, []byte{0x01, 0x02, 0x03}, false)
	assert.Equal(t, [][]byte{{0x01, 0x02, 0x03}}, recipientChannel.binaryMessages)
	assert.Equal(t, []string{"hello"}, recipientChannel.textMessages)

	// Send failures
	assert.True(t, s.isDataChannelPeerRegistered(failingRecipient))
	s.BroadcastDataChannelFrom(failingRecipient, []byte("still active"), true)
	assert.True(t, s.isDataChannelPeerRegistered(failingRecipient))
	assert.Equal(t, []string{"hello", "still active"}, recipientChannel.textMessages)

	// Unregister a peer
	s.RemoveDataChannelPeer(recipient)
	s.BroadcastDataChannelFrom(sender, []byte("after unregister"), true)
	assert.Equal(t, []string{"hello", "still active"}, recipientChannel.textMessages)

	// Replace a peer
	oldChannel := &fakeDataChannel{}
	newChannel := &fakeDataChannel{}
	oldPeer := s.AddDataChannelPeer("duplicate", oldChannel)
	newPeer := s.AddDataChannelPeer("duplicate", newChannel)

	s.RemoveDataChannelPeer(oldPeer)
	assert.True(t, s.isDataChannelPeerRegistered(newPeer))
	s.BroadcastDataChannelFrom(sender, []byte("replacement"), true)
	assert.Empty(t, oldChannel.textMessages)
	assert.Equal(t, []string{"replacement"}, newChannel.textMessages)
}

func (s *Session) isDataChannelPeerRegistered(peer *datadc.Peer) bool {
	s.DataChannelPeersLock.RLock()
	defer s.DataChannelPeersLock.RUnlock()
	return s.DataChannelPeers[peer.ID()] == peer
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
