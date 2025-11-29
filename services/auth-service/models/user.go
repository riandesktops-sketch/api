package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user account
// Indexes:
//   - email: unique index for fast lookup and prevent duplicates
//   - zodiac_sign: index for filtering users by zodiac
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email" validate:"required,email"`
	Password     string             `bson:"password" json:"-"` // Never expose in JSON
	FullName     string             `bson:"full_name" json:"full_name" validate:"required"`
	DisplayName  string             `bson:"display_name" json:"display_name"`
	DateOfBirth  time.Time          `bson:"date_of_birth" json:"date_of_birth" validate:"required"`
	Gender       string             `bson:"gender" json:"gender" validate:"required,oneof=male female other"`
	ZodiacSign   string             `bson:"zodiac_sign" json:"zodiac_sign"` // Auto-calculated, immutable
	Bio          string             `bson:"bio" json:"bio"`
	AvatarURL    string             `bson:"avatar_url" json:"avatar_url"`
	
	// Stats (denormalized for performance)
	TotalPosts   int `bson:"total_posts" json:"total_posts"`
	FriendsCount int `bson:"friends_count" json:"friends_count"`
	
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Email       string    `json:"email" validate:"required,email"`
	Password    string    `json:"password" validate:"required,min=8"`
	FullName    string    `json:"full_name" validate:"required"`
	DateOfBirth time.Time `json:"date_of_birth" validate:"required"`
	Gender      string    `json:"gender" validate:"required,oneof=male female other"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatar_url"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

// RefreshToken represents a refresh token document
// Indexes:
//   - token: unique index for fast lookup
//   - created_at: TTL index (30 days = 2592000 seconds)
type RefreshToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Token     string             `bson:"token" json:"token"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
