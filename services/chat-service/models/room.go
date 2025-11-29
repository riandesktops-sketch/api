package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Room represents a discussion room
type Room struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name" validate:"required"`
	Topic       string             `bson:"topic" json:"topic"`
	ZodiacFilter string            `bson:"zodiac_filter" json:"zodiac_filter"` // Optional filter
	CreatorID   primitive.ObjectID `bson:"creator_id" json:"creator_id"`
	MemberCount int                `bson:"member_count" json:"member_count"`
	
	// Embedded last message for preview (data locality)
	LastMessage *RoomMessagePreview `bson:"last_message" json:"last_message"`
	
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// RoomMessagePreview represents embedded last message preview
type RoomMessagePreview struct {
	Content   string    `bson:"content" json:"content"`
	SenderID  string    `bson:"sender_id" json:"sender_id"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

// RoomMessage represents a message in a room
// CRITICAL: TTL Index for storage optimization
// Indexes:
//   - created_at: TTL index (24 hours = 86400 seconds) - AUTO DELETE
//   - room_id: index for fast room message lookup
// Reference: DDIA Ch. 3 - TTL prevents storage bloat
type RoomMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID    primitive.ObjectID `bson:"room_id" json:"room_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Username  string             `bson:"username" json:"username"` // Denormalized for display
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"` // TTL index on this field
}

// CreateRoomRequest represents create room request
type CreateRoomRequest struct {
	Name         string `json:"name" validate:"required"`
	Topic        string `json:"topic"`
	ZodiacFilter string `json:"zodiac_filter"`
}

// WebSocketMessage represents WebSocket message format
type WebSocketMessage struct {
	Type      string    `json:"type"` // "message", "join", "leave"
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
