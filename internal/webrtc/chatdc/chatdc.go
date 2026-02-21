package chatdc

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/glimesh/broadcast-box/internal/chat"
	chatruntime "github.com/glimesh/broadcast-box/internal/chat/runtime"
	"github.com/pion/webrtc/v4"
)

const DataChannelLabel = "bb-chat-v1"

const (
	inboundTypeSend = "chat.send"

	outboundTypeConnected = "chat.connected"
	outboundTypeHistory   = "chat.history"
	outboundTypeMessage   = "chat.message"
	outboundTypeAck       = "chat.ack"
	outboundTypeError     = "chat.error"
)

type inboundMessage struct {
	Type          string `json:"type"`
	ClientMessage string `json:"clientMsgId,omitempty"`
	Text          string `json:"text,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
}

type outboundMessage struct {
	Type          string       `json:"type"`
	ClientMessage string       `json:"clientMsgId,omitempty"`
	Error         string       `json:"error,omitempty"`
	EventID       uint64       `json:"eventId,omitempty"`
	Message       chat.Message `json:"message,omitempty"`
	Events        []chat.Event `json:"events,omitempty"`
}

func Bind(streamKey string, peerID string, dataChannel *webrtc.DataChannel) {
	if dataChannel.Label() != DataChannelLabel {
		return
	}

	if chatruntime.ChatManager == nil {
		log.Println("ChatDC.Bind: chat manager not configured")
		return
	}

	var (
		closeSubscription func()
		closeLock         sync.Mutex
		writeLock         sync.Mutex
	)
	closeSubscription = func() {}

	runCloseSubscription := func() {
		closeLock.Lock()
		defer closeLock.Unlock()
		closeSubscription()
	}

	send := func(payload outboundMessage) bool {
		data, err := json.Marshal(payload)
		if err != nil {
			log.Println("ChatDC.Bind: marshal error", err)
			return false
		}

		writeLock.Lock()
		defer writeLock.Unlock()

		if err := dataChannel.SendText(string(data)); err != nil {
			log.Println("ChatDC.Bind: send error", err)
			return false
		}

		return true
	}

	dataChannel.OnOpen(func() {
		log.Println("ChatDC.Bind: open", streamKey, peerID)

		ch, unsubscribe, history, err := chatruntime.ChatManager.SubscribeStream(streamKey, 0)
		if err != nil {
			_ = send(outboundMessage{Type: outboundTypeError, Error: err.Error()})
			return
		}

		closeLock.Lock()
		closeSubscription = sync.OnceFunc(unsubscribe)
		closeLock.Unlock()
		if !send(outboundMessage{Type: outboundTypeConnected}) {
			runCloseSubscription()
			return
		}

		if len(history) > 0 {
			if !send(outboundMessage{Type: outboundTypeHistory, Events: history}) {
				runCloseSubscription()
				return
			}
		}

		go func() {
			for event := range ch {
				switch event.Type {
				case chat.EventTypeConnected:
					continue
				case chat.EventTypeMessage:
					if !send(outboundMessage{Type: outboundTypeMessage, EventID: event.ID, Message: event.Message}) {
						runCloseSubscription()
						return
					}
				}
			}
		}()
	})

	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		var inbound inboundMessage
		if err := json.Unmarshal(msg.Data, &inbound); err != nil {
			_ = send(outboundMessage{Type: outboundTypeError, Error: "invalid payload"})
			return
		}

		switch inbound.Type {
		case inboundTypeSend:
			text := strings.TrimSpace(inbound.Text)
			displayName := strings.TrimSpace(inbound.DisplayName)

			if len(text) < 1 || len(text) > 2000 {
				_ = send(outboundMessage{Type: outboundTypeError, Error: "invalid message length", ClientMessage: inbound.ClientMessage})
				return
			}

			if len(displayName) < 1 || len(displayName) > 80 {
				_ = send(outboundMessage{Type: outboundTypeError, Error: "invalid display name length", ClientMessage: inbound.ClientMessage})
				return
			}

			if err := chatruntime.ChatManager.SendToStream(streamKey, text, displayName); err != nil {
				_ = send(outboundMessage{Type: outboundTypeError, Error: err.Error(), ClientMessage: inbound.ClientMessage})
				return
			}

			_ = send(outboundMessage{Type: outboundTypeAck, ClientMessage: inbound.ClientMessage})
		default:
			_ = send(outboundMessage{Type: outboundTypeError, Error: "unsupported message type"})
		}
	})

	dataChannel.OnClose(func() {
		log.Println("ChatDC.Bind: closed", streamKey, peerID)
		runCloseSubscription()
	})

	dataChannel.OnError(func(err error) {
		log.Println("ChatDC.Bind: error", streamKey, peerID, err)
		runCloseSubscription()
	})
}
