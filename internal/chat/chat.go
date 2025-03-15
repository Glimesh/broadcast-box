package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// Message represents a chat message
type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Sender  string `json:"sender,omitempty"`
	NewNick string `json:"newNick,omitempty"` // Add this field for nickname changes
}

// Client represents a connected chat client
type Client struct {
	ID         string
	Conn       *websocket.Conn
	Room       *Room
	SendBuffer chan []byte
	Manager    *ChatManager
	Username   string // Add username field for tracking current nickname
}

// Room represents a chat room
type Room struct {
	Name    string
	Clients map[string]*Client
	mu      sync.RWMutex
}

// ChatManager manages all chat rooms
type ChatManager struct {
	Rooms map[string]*Room
	mu    sync.RWMutex
}

// NewChatManager creates a new chat manager
func NewChatManager() *ChatManager {
	return &ChatManager{
		Rooms: make(map[string]*Room),
	}
}

// CreateRoom creates a new chat room with the given name
func (cm *ChatManager) CreateRoom(name string) *Room {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if room already exists
	if room, exists := cm.Rooms[name]; exists {
		return room
	}

	// Create new room
	room := &Room{
		Name:    name,
		Clients: make(map[string]*Client),
	}

	cm.Rooms[name] = room
	return room
}

// GetRoom returns a room by name, or nil if it doesn't exist
func (cm *ChatManager) GetRoom(name string) *Room {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.Rooms[name]
}

// UpdateClientID updates a client's ID in the room's clients map
func (room *Room) UpdateClientID(oldID string, client *Client) {
	room.mu.Lock()
	defer room.mu.Unlock()

	// Check if the old ID exists
	if _, exists := room.Clients[oldID]; !exists {
		log.Printf("Client with ID %s not found in room %s", oldID, room.Name)
		return
	}

	// Add client with new ID
	room.Clients[client.ID] = client

	// Remove old ID entry
	delete(room.Clients, oldID)

	log.Printf("Updated client ID from %s to %s in room %s", oldID, client.ID, room.Name)
}

// AddClient adds a client to a room
func (room *Room) AddClient(client *Client) {
	room.mu.Lock()
	defer room.mu.Unlock()

	// Check if client with this ID already exists
	if _, exists := room.Clients[client.ID]; exists {
		log.Printf("Client with ID %s already exists in room %s", client.ID, room.Name)
		return
	}

	room.Clients[client.ID] = client

	// Log connection
	log.Printf("Client %s added to room %s. Total clients: %d", client.ID, room.Name, len(room.Clients))
}

// RemoveClient removes a client from a room
func (room *Room) RemoveClient(clientID string) {
	room.mu.Lock()
	defer room.mu.Unlock()

	// Check if client exists before deleting
	if _, exists := room.Clients[clientID]; !exists {
		log.Printf("Client %s not found in room %s", clientID, room.Name)
		return
	}

	delete(room.Clients, clientID)

	// Log disconnection
	log.Printf("Client %s removed from room %s. Total clients: %d", clientID, room.Name, len(room.Clients))
}

// BroadcastUserCount sends the current user count to all clients in the room
func (room *Room) BroadcastUserCount() {
	room.mu.RLock()
	count := len(room.Clients)
	clients := make(map[string]*Client, len(room.Clients))

	// Create a copy of clients map to avoid deadlock
	for id, client := range room.Clients {
		clients[id] = client
	}
	room.mu.RUnlock()

	// Log current count
	log.Printf("Broadcasting user count %d for room %s", count, room.Name)

	// Create user count message
	countMsg := Message{
		Type:    "usercount",
		Content: strconv.Itoa(count),
	}

	// Convert to JSON
	jsonMsg, err := json.Marshal(countMsg)
	if err != nil {
		log.Printf("Error marshaling user count message: %v", err)
		return
	}

	// Broadcast to all clients
	for _, client := range clients {
		select {
		case client.SendBuffer <- jsonMsg:
			// Message sent
		default:
			// Buffer full, skip
		}
	}
}

// Broadcast sends a message to all clients in the room
func (room *Room) Broadcast(message []byte, senderID string) {
	room.mu.RLock()
	defer room.mu.RUnlock()

	for id, client := range room.Clients {
		// Don't send message back to sender if senderID is provided
		if senderID == "" || id != senderID {
			select {
			case client.SendBuffer <- message:
				// Message sent
			default:
				// Buffer full, skip
			}
		}
	}
}

// Add method to update client nickname
func (client *Client) UpdateNickname(newNick string) string {
	oldNick := client.Username
	client.Username = newNick

	// Get the old client ID before changing it
	oldID := client.ID

	// Extract the remote address part from the original ID
	addrPart := client.ID[:strings.LastIndex(client.ID, "_")]

	// Create the new client ID
	newID := addrPart + "_" + newNick

	// Update client ID
	client.ID = newID

	// Update the client reference in the room's clients map
	client.Room.UpdateClientID(oldID, client)

	// Log nickname change
	log.Printf("Client nickname changed from %s to %s in room %s", oldNick, newNick, client.Room.Name)

	return oldNick
}

// Start initializes the client's goroutines for reading and writing
func (client *Client) Start() {
	// Add client to room
	client.Room.AddClient(client)

	// Broadcast updated user count
	client.Room.BroadcastUserCount()

	// Start goroutines for reading and writing
	go client.readPump()
	go client.writePump()
}

// readPump reads messages from the WebSocket
func (client *Client) readPump() {
	defer func() {
		client.Conn.Close()
		client.Room.RemoveClient(client.ID)

		// Broadcast updated user count
		client.Room.BroadcastUserCount()

		// Close send buffer
		close(client.SendBuffer)
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error: %v", err)
			}
			break
		}

		// Try to parse the message
		var msgObj Message
		if err := json.Unmarshal(message, &msgObj); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// Process message based on type
		switch msgObj.Type {
		case "message":
			// Regular chat message - broadcast to everyone in the room except sender
			// No need to echo messages back to sender as they're handled locally
			client.Room.Broadcast(message, client.ID)
		case "join", "leave":
			// System messages for join/leave events - broadcast to everyone
			client.Room.Broadcast(message, client.ID)
			// No need to echo these back to sender
		case "nickname":
			// Nickname change message
			oldNick := client.UpdateNickname(msgObj.NewNick)

			// Create system message about nickname change
			nickChangeMsg := Message{
				Type:    "system",
				Content: oldNick + " changed their nickname to " + msgObj.NewNick,
			}

			// Marshal the message
			jsonMsg, err := json.Marshal(nickChangeMsg)
			if err != nil {
				log.Printf("Error marshaling nickname change message: %v", err)
				continue
			}

			// Broadcast to everyone including the sender
			client.Room.Broadcast(jsonMsg, "")

		default:
			// Unknown message type
			log.Printf("Unknown message type: %s", msgObj.Type)
		}
	}
}

// writePump writes messages to the WebSocket
func (client *Client) writePump() {
	defer client.Conn.Close()

	for {
		message, ok := <-client.SendBuffer
		if !ok {
			// Channel closed
			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error writing message: %v", err)
			return
		}
	}
}

// Upgrader for WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for now
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ChatWebSocketHandler handles WebSocket connections for chat
func ChatWebSocketHandler(w http.ResponseWriter, r *http.Request, cm *ChatManager) {
	// Get room name from query parameter
	roomName := r.URL.Query().Get("room")
	if roomName == "" {
		http.Error(w, "Room name is required", http.StatusBadRequest)
		return
	}

	// Get username from query parameter
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "anonymous"
	}

	// Get or create room
	room := cm.GetRoom(roomName)
	if room == nil {
		room = cm.CreateRoom(roomName)
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	// Create a unique ID using remote address and username
	clientID := conn.RemoteAddr().String() + "_" + username

	// Create client
	client := &Client{
		ID:         clientID,
		Conn:       conn,
		Room:       room,
		SendBuffer: make(chan []byte, 256),
		Manager:    cm,
		Username:   username, // Store the username in the client
	}

	// Start client goroutines
	client.Start()
}
