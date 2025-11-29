package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatSession represents an AI chat session
type ChatSession struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Title     string             `bson:"title" json:"title"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// MessageSender represents who sent the message
type MessageSender string

const (
	SenderUser MessageSender = "USER"
	SenderAI   MessageSender = "AI"
)

// Message represents a chat message
// CRITICAL: TTL Index for storage optimization
// Indexes:
//   - created_at: TTL index (48 hours = 172800 seconds) - AUTO DELETE
//   - session_id: index for fast session message lookup
// Reference: DDIA Ch. 3 - TTL prevents storage bloat
type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SessionID primitive.ObjectID `bson:"session_id" json:"session_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Sender    MessageSender      `bson:"sender" json:"sender"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"` // TTL index on this field
}

// SendMessageRequest represents send message request
type SendMessageRequest struct {
	Message string `json:"message" validate:"required,min=1"`
}

// MessageResponse represents message response
type MessageResponse struct {
	UserMessage *Message `json:"user_message"`
	AIMessage   *Message `json:"ai_message"`
}
