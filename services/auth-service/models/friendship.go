package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FriendshipStatus represents the status of a friendship
type FriendshipStatus string

const (
	StatusPending  FriendshipStatus = "PENDING"
	StatusAccepted FriendshipStatus = "ACCEPTED"
	StatusRejected FriendshipStatus = "REJECTED"
)

// Friendship represents denormalized friendship graph
// Reference: CLRS Ch. 22 - Adjacency list representation for O(1) lookup
// Indexes:
//   - user_id: index for finding user's friendship document
//   - {user_id: 1, friend_ids: 1}: compound index for O(1) friendship check
type Friendship struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID   `bson:"user_id" json:"user_id"`
	FriendIDs       []primitive.ObjectID `bson:"friend_ids" json:"friend_ids"`           // Array of accepted friends
	PendingSent     []primitive.ObjectID `bson:"pending_sent" json:"pending_sent"`       // Requests sent by user
	PendingReceived []primitive.ObjectID `bson:"pending_received" json:"pending_received"` // Requests received by user
	UpdatedAt       time.Time            `bson:"updated_at" json:"updated_at"`
}

// FriendRequest represents a friend request (for transaction tracking)
type FriendRequest struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SenderID   primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	ReceiverID primitive.ObjectID `bson:"receiver_id" json:"receiver_id"`
	Status     FriendshipStatus   `bson:"status" json:"status"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

// SendFriendRequestInput represents friend request input
type SendFriendRequestInput struct {
	TargetUserID string `json:"target_user_id" validate:"required"`
}

// AcceptRejectRequestInput represents accept/reject input
type AcceptRejectRequestInput struct {
	Action string `json:"action" validate:"required,oneof=accept reject"`
}

// FriendStatusResponse represents friendship status response
type FriendStatusResponse struct {
	Status string `json:"status"` // ARE_FRIENDS, PENDING, NOT_FRIENDS
}
