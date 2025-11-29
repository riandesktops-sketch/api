package websocket

import (
	"log"
	"sync"

	"zodiac-ai-backend/services/chat-service/models"

	"github.com/gofiber/websocket/v2"
)

// Client represents a WebSocket client
type Client struct {
	ID       string
	RoomID   string
	UserID   string
	Username string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan *models.WebSocketMessage
}

// Hub manages WebSocket connections and rooms
// Reference: Go Concurrency Patterns - Hub pattern with channels
type Hub struct {
	// Registered clients per room
	rooms map[string]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to room
	broadcast chan *BroadcastMessage

	// Mutex for thread-safe room access
	mu sync.RWMutex
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	RoomID  string
	Message *models.WebSocketMessage
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

// Run starts the hub's main loop
// Handles register, unregister, and broadcast events using Go channels
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[client.RoomID]; !ok {
				h.rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.rooms[client.RoomID][client] = true
			h.mu.Unlock()

			log.Printf("Client %s joined room %s", client.Username, client.RoomID)

			// Broadcast join message
			joinMsg := &models.WebSocketMessage{
				Type:      "join",
				UserID:    client.UserID,
				Username:  client.Username,
				Content:   client.Username + " joined the room",
				Timestamp: getCurrentTime(),
			}
			h.broadcastToRoom(client.RoomID, joinMsg)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.RoomID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)

					// Remove room if empty
					if len(clients) == 0 {
						delete(h.rooms, client.RoomID)
					}
				}
			}
			h.mu.Unlock()

			log.Printf("Client %s left room %s", client.Username, client.RoomID)

			// Broadcast leave message
			leaveMsg := &models.WebSocketMessage{
				Type:      "leave",
				UserID:    client.UserID,
				Username:  client.Username,
				Content:   client.Username + " left the room",
				Timestamp: getCurrentTime(),
			}
			h.broadcastToRoom(client.RoomID, leaveMsg)

		case broadcastMsg := <-h.broadcast:
			h.broadcastToRoom(broadcastMsg.RoomID, broadcastMsg.Message)
		}
	}
}

// broadcastToRoom sends message to all clients in a room
func (h *Hub) broadcastToRoom(roomID string, message *models.WebSocketMessage) {
	h.mu.RLock()
	clients, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for client := range clients {
		select {
		case client.Send <- message:
		default:
			// Client's send channel is full, close and unregister
			close(client.Send)
			h.mu.Lock()
			delete(clients, client)
			h.mu.Unlock()
		}
	}
}

// BroadcastMessage broadcasts a message to a room
func (h *Hub) BroadcastMessage(roomID string, message *models.WebSocketMessage) {
	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: message,
	}
}

// GetRoomClients gets number of clients in a room
func (h *Hub) GetRoomClients(roomID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[roomID]; ok {
		return len(clients)
	}
	return 0
}

// ReadPump reads messages from WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		var msg models.WebSocketMessage
		if err := c.Conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Set metadata
		msg.UserID = c.UserID
		msg.Username = c.Username
		msg.Timestamp = getCurrentTime()

		// Broadcast to room
		c.Hub.BroadcastMessage(c.RoomID, &msg)
	}
}

// WritePump writes messages to WebSocket connection
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		if err := c.Conn.WriteJSON(message); err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
}

// getCurrentTime returns current timestamp
func getCurrentTime() string {
	return ""
}
