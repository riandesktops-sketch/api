package handlers

import (
	"zodiac-ai-backend/pkg/middleware"
	"zodiac-ai-backend/pkg/response"
	"zodiac-ai-backend/services/chat-service/models"
	"zodiac-ai-backend/services/chat-service/repositories"
	"zodiac-ai-backend/services/chat-service/websocket"

	"github.com/gofiber/fiber/v2"
	ws "github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RoomHandler handles room HTTP and WebSocket requests
type RoomHandler struct {
	roomRepo *repositories.RoomRepository
	hub      *websocket.Hub
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(roomRepo *repositories.RoomRepository, hub *websocket.Hub) *RoomHandler {
	return &RoomHandler{
		roomRepo: roomRepo,
		hub:      hub,
	}
}

// CreateRoom creates a new discussion room
// POST /rooms
func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req models.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	userObjID, _ := primitive.ObjectIDFromHex(userID)

	room := &models.Room{
		Name:         req.Name,
		Topic:        req.Topic,
		ZodiacFilter: req.ZodiacFilter,
		CreatorID:    userObjID,
	}

	if err := h.roomRepo.Create(c.Context(), room); err != nil {
		return response.InternalServerError(c, "Failed to create room")
	}

	return response.Created(c, "Room created successfully", room)
}

// GetRooms gets list of rooms
// GET /rooms
func (h *RoomHandler) GetRooms(c *fiber.Ctx) error {
	topic := c.Query("topic", "")
	zodiacFilter := c.Query("zodiac", "")
	limit := c.QueryInt("limit", 20)

	rooms, err := h.roomRepo.GetRooms(c.Context(), topic, zodiacFilter, limit)
	if err != nil {
		return response.InternalServerError(c, "Failed to get rooms")
	}

	return response.Success(c, "Rooms retrieved successfully", rooms)
}

// JoinRoom handles WebSocket connection to a room
// WS /rooms/:id/ws
func (h *RoomHandler) JoinRoom(c *ws.Conn) {
	// Get room ID from params
	roomID := c.Params("id")
	if roomID == "" {
		c.Close()
		return
	}

	// Get user info from locals (set by middleware)
	userID := c.Locals("user_id")
	username := c.Locals("username")

	if userID == nil || username == nil {
		c.Close()
		return
	}

	// Create client
	client := &websocket.Client{
		ID:       userID.(string) + "_" + roomID,
		RoomID:   roomID,
		UserID:   userID.(string),
		Username: username.(string),
		Conn:     c,
		Hub:      h.hub,
		Send:     make(chan *models.WebSocketMessage, 256),
	}

	// Register client
	h.hub.register <- client

	// Start read and write pumps
	go client.WritePump()
	client.ReadPump()
}
