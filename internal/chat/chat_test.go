package chat

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestChatRoomCreation(t *testing.T) {
	chatManager := NewChatManager()
	
	roomName := "test-room"
	room := chatManager.CreateRoom(roomName)
	
	if room == nil {
		t.Fatalf("Failed to create chat room")
	}
	
	if room.Name != roomName {
		t.Errorf("Room has incorrect name: got %s, want %s", room.Name, roomName)
	}
	
	if len(room.Clients) != 0 {
		t.Errorf("New room should have 0 clients, got %d", len(room.Clients))
	}
}

func TestWebSocketHandler(t *testing.T) {
	chatManager := NewChatManager()
	chatManager.CreateRoom("test-room")
	
	// Create a WebSocket server using the handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ChatWebSocketHandler(w, r, chatManager)
	}))
	defer server.Close()
	
	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?room=test-room"
	
	// Connect a WebSocket client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Could not connect to WebSocket: %v", err)
	}
	defer ws.Close()
	
	// Send a message
	testMessage := []byte(`{"type": "message", "content": "Hello, World!"}`)
	if err := ws.WriteMessage(websocket.TextMessage, testMessage); err != nil {
		t.Fatalf("Could not send message: %v", err)
	}
	
	// Should receive the same message back (echo)
	_, receivedMsg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read message: %v", err)
	}
	
	if string(receivedMsg) != string(testMessage) {
		t.Errorf("Received incorrect message: got %s, want %s", receivedMsg, testMessage)
	}
}

func TestMessageBroadcast(t *testing.T) {
	chatManager := NewChatManager()
	chatManager.CreateRoom("test-room")
	
	// Create a WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ChatWebSocketHandler(w, r, chatManager)
	}))
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?room=test-room"
	
	// Connect first client
	ws1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Could not connect first client: %v", err)
	}
	defer ws1.Close()
	
	// Connect second client
	ws2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Could not connect second client: %v", err)
	}
	defer ws2.Close()
	
	// Send a message from first client
	testMessage := []byte(`{"type": "message", "content": "Hello from client 1"}`)
	if err := ws1.WriteMessage(websocket.TextMessage, testMessage); err != nil {
		t.Fatalf("Could not send message: %v", err)
	}
	
	// Second client should receive the message
	_, receivedMsg, err := ws2.ReadMessage()
	if err != nil {
		t.Fatalf("Second client could not read message: %v", err)
	}
	
	if string(receivedMsg) != string(testMessage) {
		t.Errorf("Second client received incorrect message: got %s, want %s", receivedMsg, testMessage)
	}
}

func TestJoinLeaveMessages(t *testing.T) {
	chatManager := NewChatManager()
	chatManager.CreateRoom("test-room")
	
	// Create a WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ChatWebSocketHandler(w, r, chatManager)
	}))
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?room=test-room"
	
	// Connect first client
	ws1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Could not connect first client: %v", err)
	}
	defer ws1.Close()
	
	// Connect second client
	ws2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Could not connect second client: %v", err)
	}
	defer ws2.Close()
	
	// Send a join message from first client
	joinMessage := []byte(`{"type": "join", "sender": "user1", "content": "joined the chat"}`)
	if err := ws1.WriteMessage(websocket.TextMessage, joinMessage); err != nil {
		t.Fatalf("Could not send join message: %v", err)
	}
	
	// Second client should receive the join message
	_, receivedJoin, err := ws2.ReadMessage()
	if err != nil {
		t.Fatalf("Second client could not read join message: %v", err)
	}
	
	if string(receivedJoin) != string(joinMessage) {
		t.Errorf("Second client received incorrect join message: got %s, want %s", receivedJoin, joinMessage)
	}
	
	// Send a leave message from first client
	leaveMessage := []byte(`{"type": "leave", "sender": "user1", "content": "left the chat"}`)
	if err := ws1.WriteMessage(websocket.TextMessage, leaveMessage); err != nil {
		t.Fatalf("Could not send leave message: %v", err)
	}
	
	// Second client should receive the leave message
	_, receivedLeave, err := ws2.ReadMessage()
	if err != nil {
		t.Fatalf("Second client could not read leave message: %v", err)
	}
	
	if string(receivedLeave) != string(leaveMessage) {
		t.Errorf("Second client received incorrect leave message: got %s, want %s", receivedLeave, leaveMessage)
	}
}