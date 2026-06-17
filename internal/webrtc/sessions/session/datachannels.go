package session

import (
	"log/slog"
	"sync"

	"github.com/glimesh/broadcast-box/internal/webrtc/datadc"
)

func (s *Session) AddDataChannelPeer(peerID string, channel datadc.Sender) *datadc.Peer {
	s.DataChannelPeersLock.Lock()
	defer s.DataChannelPeersLock.Unlock()

	if s.DataChannelPeers == nil {
		s.DataChannelPeers = map[string]*datadc.Peer{}
	}

	peer := datadc.NewPeer(peerID, channel)
	s.DataChannelPeers[peerID] = peer
	return peer
}

func (s *Session) RemoveDataChannelPeer(peer *datadc.Peer) {
	if peer == nil {
		return
	}

	s.DataChannelPeersLock.Lock()
	defer s.DataChannelPeersLock.Unlock()

	if s.DataChannelPeers[peer.ID()] == peer {
		delete(s.DataChannelPeers, peer.ID())
	}
}

func (s *Session) BroadcastDataChannelFrom(sender *datadc.Peer, payload []byte, isString bool) {
	if sender == nil {
		return
	}

	s.DataChannelPeersLock.RLock()
	if s.DataChannelPeers[sender.ID()] != sender {
		s.DataChannelPeersLock.RUnlock()
		return
	}

	recipients := make([]*datadc.Peer, 0, len(s.DataChannelPeers))
	for peerID, peer := range s.DataChannelPeers {
		if peerID != sender.ID() {
			recipients = append(recipients, peer)
		}
	}
	s.DataChannelPeersLock.RUnlock()

	var wg sync.WaitGroup
	for _, recipient := range recipients {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := recipient.Send(payload, isString); err != nil {
				slog.Error(
					"DataDC.Broadcast: send error",
					"streamKey", s.StreamKey,
					"senderPeerID", sender.ID(),
					"recipientPeerID", recipient.ID(),
					"err", err,
				)
			}
		}()
	}
	wg.Wait()
}
