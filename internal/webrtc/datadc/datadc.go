package datadc

import (
	"log/slog"
	"sync"

	"github.com/pion/webrtc/v4"
)

const DataChannelLabel = "bb-data-v1"

type Handler struct {
	manager *Manager
}

func NewHandler(dm *Manager) *Handler {
	return &Handler{manager: dm}
}

type Manager struct {
	channelsLock sync.RWMutex
	channels     map[string]map[string]*peerChannel
}

func NewManager() *Manager {
	return &Manager{
		channels: map[string]map[string]*peerChannel{},
	}
}

type dataChannelSender interface {
	Send(data []byte) error
	SendText(s string) error
}

type peerChannel struct {
	streamKey string
	peerID    string
	channel   dataChannelSender

	writeLock sync.Mutex
}

func (h *Handler) Bind(streamKey string, peerID string, dataChannel *webrtc.DataChannel) {
	if dataChannel.Label() != DataChannelLabel {
		return
	}

	if h.manager == nil {
		slog.Info("DataDC.Bind: data manager not configured")
		return
	}

	var (
		registeredPeer     *peerChannel
		registeredPeerLock sync.Mutex
		isClosed           bool
	)

	ensureRegistered := func() *peerChannel {
		registeredPeerLock.Lock()
		defer registeredPeerLock.Unlock()

		if isClosed {
			return nil
		}

		if registeredPeer == nil {
			registeredPeer = h.manager.register(streamKey, peerID, dataChannel)
		}

		return registeredPeer
	}

	dataChannel.OnOpen(func() {
		slog.Info("DataDC.Bind: open", "streamKey", streamKey, "peerID", peerID)
		ensureRegistered()
	})

	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		h.manager.broadcastFrom(ensureRegistered(), msg.Data, msg.IsString)
	})

	dataChannel.OnClose(func() {
		slog.Info("DataDC.Bind: closed", "streamKey", streamKey, "peerID", peerID)

		registeredPeerLock.Lock()
		isClosed = true
		peer := registeredPeer
		registeredPeerLock.Unlock()
		h.manager.unregister(peer)
	})

	dataChannel.OnError(func(err error) {
		registeredPeerLock.Lock()
		wasClosed := isClosed
		isClosed = true
		peer := registeredPeer
		registeredPeerLock.Unlock()

		if wasClosed {
			return
		}

		slog.Error("DataDC.Bind: error", "streamKey", streamKey, "peerID", peerID, "err", err)
		h.manager.unregister(peer)
	})
}

func (m *Manager) register(streamKey string, peerID string, channel dataChannelSender) *peerChannel {
	if m == nil {
		return nil
	}

	peer := &peerChannel{
		streamKey: streamKey,
		peerID:    peerID,
		channel:   channel,
	}

	m.channelsLock.Lock()
	defer m.channelsLock.Unlock()

	streamChannels, ok := m.channels[streamKey]
	if !ok {
		streamChannels = map[string]*peerChannel{}
		m.channels[streamKey] = streamChannels
	}
	streamChannels[peerID] = peer

	return peer
}

func (m *Manager) unregister(peer *peerChannel) {
	if m == nil || peer == nil {
		return
	}

	m.channelsLock.Lock()
	defer m.channelsLock.Unlock()

	streamChannels, ok := m.channels[peer.streamKey]
	if !ok {
		return
	}

	if streamChannels[peer.peerID] != peer {
		return
	}

	delete(streamChannels, peer.peerID)
	if len(streamChannels) == 0 {
		delete(m.channels, peer.streamKey)
	}
}

func (m *Manager) broadcastFrom(sender *peerChannel, payload []byte, isString bool) {
	if m == nil || sender == nil || !m.isRegistered(sender) {
		return
	}

	recipients := m.snapshotRecipients(sender.streamKey, sender.peerID)

	for _, recipient := range recipients {
		if err := recipient.send(payload, isString); err != nil {
			slog.Error(
				"DataDC.Broadcast: send error",
				"streamKey", sender.streamKey,
				"senderPeerID", sender.peerID,
				"recipientPeerID", recipient.peerID,
				"err", err,
			)
		}
	}
}

func (m *Manager) isRegistered(peer *peerChannel) bool {
	m.channelsLock.RLock()
	defer m.channelsLock.RUnlock()

	streamChannels, ok := m.channels[peer.streamKey]
	if !ok {
		return false
	}

	return streamChannels[peer.peerID] == peer
}

func (m *Manager) snapshotRecipients(streamKey string, senderPeerID string) []*peerChannel {
	if m == nil {
		return nil
	}

	m.channelsLock.RLock()
	defer m.channelsLock.RUnlock()

	streamChannels, ok := m.channels[streamKey]
	if !ok {
		return nil
	}

	recipients := make([]*peerChannel, 0, len(streamChannels))
	for peerID, peer := range streamChannels {
		if peerID == senderPeerID {
			continue
		}
		recipients = append(recipients, peer)
	}

	return recipients
}

func (p *peerChannel) send(payload []byte, isString bool) error {
	p.writeLock.Lock()
	defer p.writeLock.Unlock()

	if isString {
		return p.channel.SendText(string(payload))
	}

	return p.channel.Send(payload)
}
