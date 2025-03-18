package chat

import (
	"encoding/json"
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

// readUntilMessageType reads WebSocket messages until it finds one with the specified type
func readUntilMessageType(t *testing.T, ws *websocket.Conn, messageType string) (Message, error) {
	var msg Message
	
	// Try a few times to find the message type
	for i := 0; i < 5; i++ {
		_, receivedMsg, err := ws.ReadMessage()
		if err != nil {
			return msg, err
		}
		
		if err := json.Unmarshal(receivedMsg, &msg); err != nil {
			return msg, err
		}
		
		if msg.Type == messageType {
			return msg, nil
		}
	}
	
	t.Logf("Could not find message of type '%s' after multiple reads", messageType)
	return msg, nil
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
	
	// First message will be usercount - read and discard
	_, _, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read initial usercount message: %v", err)
	}
	
	// There will also be a userlist message - read and discard
	_, _, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read initial userlist message: %v", err)
	}
	
	// Send a message
	testMessage := []byte(`{"type": "message", "content": "Hello, World!"}`)
	if err := ws.WriteMessage(websocket.TextMessage, testMessage); err != nil {
		t.Fatalf("Could not send message: %v", err)
	}
	
	// We don't receive echo messages anymore - they're only sent to other clients
	// This test is now validating connection can be established
	t.Log("WebSocket connection established successfully")
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
	
	// First client will receive usercount and userlist messages - read and discard both
	_, _, err = ws1.ReadMessage()
	if err != nil {
		t.Fatalf("First client could not read initial usercount message: %v", err)
	}
	_, _, err = ws1.ReadMessage()
	if err != nil {
		t.Fatalf("First client could not read initial userlist message: %v", err)
	}
	
	// Connect second client
	ws2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Could not connect second client: %v", err)
	}
	defer ws2.Close()
	
	// Second client will receive usercount and userlist messages - read and discard both
	_, _, err = ws2.ReadMessage()
	if err != nil {
		t.Fatalf("Second client could not read initial usercount message: %v", err)
	}
	_, _, err = ws2.ReadMessage()
	if err != nil {
		t.Fatalf("Second client could not read initial userlist message: %v", err)
	}
	
	// First client will also receive the updated usercount and userlist - read and discard both
	_, _, err = ws1.ReadMessage()
	if err != nil {
		t.Fatalf("First client could not read updated usercount message: %v", err)
	}
	_, _, err = ws1.ReadMessage()
	if err != nil {
		t.Fatalf("First client could not read updated userlist message: %v", err)
	}
	
	// Send a message from first client
	testMessage := []byte(`{"type": "message", "content": "Hello from client 1"}`)
	if err := ws1.WriteMessage(websocket.TextMessage, testMessage); err != nil {
		t.Fatalf("Could not send message: %v", err)
	}
	
	// Second client should receive the message, but may need to skip userlist messages
	msg, err := readUntilMessageType(t, ws2, "message")
	if err != nil {
		t.Fatalf("Second client could not read message: %v", err)
	}
	
	if msg.Type != "message" || msg.Content != "Hello from client 1" {
		t.Errorf("Second client received incorrect message: got %+v, want message with content 'Hello from client 1'", msg)
	} else {
		t.Log("Message broadcast test passed successfully")
	}
}

func TestUserListUpdates(t *testing.T) {
	chatManager := NewChatManager()
	chatManager.CreateRoom("test-room")
	
	// Create a WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ChatWebSocketHandler(w, r, chatManager)
	}))
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?room=test-room&username=user1"
	
	// Connect first client with username user1
	ws1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Could not connect first client: %v", err)
	}
	defer ws1.Close()
	
	// First client will receive usercount and userlist messages
	_, _, err = ws1.ReadMessage() // skip usercount
	if err != nil {
		t.Fatalf("First client could not read initial usercount message: %v", err)
	}
	
	// Read userlist message
	var userListMsg Message
	_, rawUserList, err := ws1.ReadMessage()
	if err != nil {
		t.Fatalf("First client could not read initial userlist message: %v", err)
	}
	
	if err := json.Unmarshal(rawUserList, &userListMsg); err != nil {
		t.Fatalf("Could not parse userlist message: %v", err)
	}
	
	if userListMsg.Type != "userlist" {
		t.Fatalf("Expected userlist message, got %s", userListMsg.Type)
	}
	
	if len(userListMsg.Users) != 1 {
		t.Errorf("Initial userlist should have 1 user, got %d", len(userListMsg.Users))
	} else if userListMsg.Users[0].Username != "user1" {
		t.Errorf("Expected username 'user1', got '%s'", userListMsg.Users[0].Username)
	} else if userListMsg.Users[0].Status != "connected" {
		t.Errorf("Expected status 'connected', got '%s'", userListMsg.Users[0].Status)
	}
	
	// Connect second client with username user2
	wsURL2 := "ws" + strings.TrimPrefix(server.URL, "http") + "?room=test-room&username=user2"
	ws2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
	if err != nil {
		t.Fatalf("Could not connect second client: %v", err)
	}
	defer ws2.Close()
	
	// Skip messages for second client
	_, _, err = ws2.ReadMessage() // skip usercount
	if err != nil {
		t.Fatalf("Second client could not read initial usercount message: %v", err)
	}
	_, _, err = ws2.ReadMessage() // skip userlist
	if err != nil {
		t.Fatalf("Second client could not read initial userlist message: %v", err)
	}
	
	// First client should receive updated usercount and userlist
	_, _, err = ws1.ReadMessage() // skip usercount
	if err != nil {
		t.Fatalf("First client could not read updated usercount message: %v", err)
	}
	
	// Read updated userlist message
	_, rawUpdatedUserList, err := ws1.ReadMessage()
	if err != nil {
		t.Fatalf("First client could not read updated userlist message: %v", err)
	}
	
	var updatedUserListMsg Message
	if err := json.Unmarshal(rawUpdatedUserList, &updatedUserListMsg); err != nil {
		t.Fatalf("Could not parse updated userlist message: %v", err)
	}
	
	if updatedUserListMsg.Type != "userlist" {
		t.Fatalf("Expected userlist message, got %s", updatedUserListMsg.Type)
	}
	
	if len(updatedUserListMsg.Users) != 2 {
		t.Errorf("Updated userlist should have 2 users, got %d", len(updatedUserListMsg.Users))
	} else {
		t.Log("User list updates correctly with multiple users")
	}
	
	// Verify both users are in the list
	foundUser1 := false
	foundUser2 := false
	
	for _, user := range updatedUserListMsg.Users {
		if user.Username == "user1" {
			foundUser1 = true
		}
		if user.Username == "user2" {
			foundUser2 = true
		}
	}
	
	if !foundUser1 || !foundUser2 {
		t.Errorf("User list missing expected users: user1=%v, user2=%v", foundUser1, foundUser2)
	}
}

func TestNicknameChange(t *testing.T) {
	chatManager := NewChatManager()
	chatManager.CreateRoom("test-room")
	
	// Create a WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ChatWebSocketHandler(w, r, chatManager)
	}))
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?room=test-room&username=oldname"
	
	// Connect client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Could not connect client: %v", err)
	}
	defer ws.Close()
	
	// Client will receive usercount and userlist messages - read and discard
	_, _, err = ws.ReadMessage() // skip usercount
	if err != nil {
		t.Fatalf("Client could not read initial usercount message: %v", err)
	}
	_, _, err = ws.ReadMessage() // skip userlist
	if err != nil {
		t.Fatalf("Client could not read initial userlist message: %v", err)
	}
	
	// Send a nickname change message
	nickChangeMessage := []byte(`{"type": "nickname", "newNick": "newname"}`)
	if err := ws.WriteMessage(websocket.TextMessage, nickChangeMessage); err != nil {
		t.Fatalf("Could not send nickname change message: %v", err)
	}
	
	// Client should receive system message about nickname change and updated userlist
	var sysMsg Message
	var userlistMsg Message
	var foundSystemMsg bool
	var foundUserlistMsg bool
	
	// Try to read both messages in any order
	for i := 0; i < 5 && (!foundSystemMsg || !foundUserlistMsg); i++ {
		_, receivedMsg, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("Could not read message: %v", err)
		}
		
		var msg Message
		if err := json.Unmarshal(receivedMsg, &msg); err != nil {
			t.Fatalf("Could not parse message: %v", err)
		}
		
		if msg.Type == "system" {
			sysMsg = msg
			foundSystemMsg = true
		} else if msg.Type == "userlist" {
			userlistMsg = msg
			foundUserlistMsg = true
		}
	}
	
	// Verify the system message
	if !foundSystemMsg {
		t.Errorf("No system message received for nickname change")
	} else if !strings.Contains(sysMsg.Content, "changed their nickname") {
		t.Errorf("System message does not contain expected text: %s", sysMsg.Content)
	}
	
	// Verify the userlist contains the new nickname
	if !foundUserlistMsg {
		t.Errorf("No userlist message received after nickname change")
	} else {
		foundNewNick := false
		for _, user := range userlistMsg.Users {
			if user.Username == "newname" {
				foundNewNick = true
				break
			}
		}
		
		if !foundNewNick {
			t.Errorf("New nickname not found in userlist after change")
		}
	}
}

func TestPingPong(t *testing.T) {
	chatManager := NewChatManager()
	chatManager.CreateRoom("test-room")
	
	// Create a WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ChatWebSocketHandler(w, r, chatManager)
	}))
	defer server.Close()
	
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?room=test-room"
	
	// Connect client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Could not connect to WebSocket: %v", err)
	}
	defer ws.Close()
	
	// First message will be usercount - read and discard
	_, _, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read initial usercount message: %v", err)
	}
	
	// Read and discard userlist message
	_, _, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read initial userlist message: %v", err)
	}
	
	// Send a ping message
	pingMessage := []byte(`{"type": "ping"}`)
	if err := ws.WriteMessage(websocket.TextMessage, pingMessage); err != nil {
		t.Fatalf("Could not send ping message: %v", err)
	}
	
	// Should receive a pong response
	pongMsg, err := readUntilMessageType(t, ws, "pong")
	if err != nil {
		t.Fatalf("Could not read pong message: %v", err)
	}
	
	if pongMsg.Type != "pong" {
		t.Errorf("Expected pong message, got %s", pongMsg.Type)
	} else {
		t.Log("Received pong response successfully")
	}
	
	// Test if client stays connected after sending messages
	testMessage := []byte(`{"type": "message", "content": "Still connected"}`)
	if err := ws.WriteMessage(websocket.TextMessage, testMessage); err != nil {
		t.Fatalf("Client disconnected unexpectedly: %v", err)
	}
	
	t.Log("Client remained connected after ping/pong exchange")
}