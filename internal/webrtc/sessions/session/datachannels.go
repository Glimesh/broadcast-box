package session

import (
	"log/slog"
	"sync"

	"github.com/glimesh/broadcast-box/internal/webrtc/chatdc"
	"github.com/pion/webrtc/v4"
)

const dataChannelLabel = "bb-data-v1"

type dataChannelSender interface {
	Send(data []byte) error
	SendText(s string) error
}

type dataChannelPeer struct {
	id      string
	channel dataChannelSender

	writeLock sync.Mutex
}

func (s *Session) registerDataChannelHandlers(peerConnection *webrtc.PeerConnection, peerID string) {
	peerConnection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		chatdc.NewHandler(s.ChatManager).Bind(s.StreamKey, peerID, dataChannel)
		s.bindDataChannel(peerID, dataChannel)
	})
}

func (s *Session) bindDataChannel(peerID string, dataChannel *webrtc.DataChannel) {
	if dataChannel.Label() != dataChannelLabel {
		return
	}

	var (
		peer     *dataChannelPeer
		peerLock sync.Mutex
		isClosed bool
	)

	register := func() *dataChannelPeer {
		peerLock.Lock()
		defer peerLock.Unlock()
		if isClosed {
			return nil
		}

		if peer == nil {
			peer = s.addDataChannelPeer(peerID, dataChannel)
		}
		return peer
	}

	closePeer := func() (*dataChannelPeer, bool) {
		peerLock.Lock()
		defer peerLock.Unlock()
		if isClosed {
			return nil, false
		}

		isClosed = true
		return peer, true
	}

	dataChannel.OnOpen(func() {
		slog.Info("DataDC.Bind: open", "streamKey", s.StreamKey, "peerID", peerID)
		register()
	})
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		s.broadcastDataChannelFrom(register(), msg.Data, msg.IsString)
	})
	dataChannel.OnClose(func() {
		slog.Info("DataDC.Bind: closed", "streamKey", s.StreamKey, "peerID", peerID)
		peer, _ := closePeer()
		s.removeDataChannelPeer(peer)
	})
	dataChannel.OnError(func(err error) {
		peer, didClose := closePeer()
		if didClose {
			slog.Error("DataDC.Bind: error", "streamKey", s.StreamKey, "peerID", peerID, "err", err)
			s.removeDataChannelPeer(peer)
		}
	})
}

func (s *Session) addDataChannelPeer(peerID string, channel dataChannelSender) *dataChannelPeer {
	s.dataChannelPeersLock.Lock()
	defer s.dataChannelPeersLock.Unlock()

	if s.dataChannelPeers == nil {
		s.dataChannelPeers = map[string]*dataChannelPeer{}
	}

	peer := &dataChannelPeer{id: peerID, channel: channel}
	s.dataChannelPeers[peerID] = peer
	return peer
}

func (s *Session) removeDataChannelPeer(peer *dataChannelPeer) {
	if peer == nil {
		return
	}

	s.dataChannelPeersLock.Lock()
	defer s.dataChannelPeersLock.Unlock()

	if s.dataChannelPeers[peer.id] == peer {
		delete(s.dataChannelPeers, peer.id)
	}
}

func (s *Session) broadcastDataChannelFrom(sender *dataChannelPeer, payload []byte, isString bool) {
	if sender == nil {
		return
	}

	s.dataChannelPeersLock.RLock()
	if s.dataChannelPeers[sender.id] != sender {
		s.dataChannelPeersLock.RUnlock()
		return
	}

	recipients := make([]*dataChannelPeer, 0, len(s.dataChannelPeers))
	for peerID, peer := range s.dataChannelPeers {
		if peerID != sender.id {
			recipients = append(recipients, peer)
		}
	}
	s.dataChannelPeersLock.RUnlock()

	var wg sync.WaitGroup
	for _, recipient := range recipients {
		wg.Go(func() {
			if err := recipient.send(payload, isString); err != nil {
				slog.Error(
					"DataDC.Broadcast: send error",
					"streamKey", s.StreamKey,
					"senderPeerID", sender.id,
					"recipientPeerID", recipient.id,
					"err", err,
				)
			}
		})
	}
	wg.Wait()
}

func (p *dataChannelPeer) send(payload []byte, isString bool) error {
	p.writeLock.Lock()
	defer p.writeLock.Unlock()

	if isString {
		return p.channel.SendText(string(payload))
	}

	return p.channel.Send(payload)
}
