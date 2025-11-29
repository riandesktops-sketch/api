package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostStatus represents post status
type PostStatus string

const (
	StatusDraft     PostStatus = "DRAFT"
	StatusPublished PostStatus = "PUBLISHED"
)

// Post represents a social feed post (insight from AI chat)
// NO TTL - Posts are permanent
// Indexes:
//   - {created_at: -1, likes_count: -1}: compound index for feed sorting
//   - author_zodiac: index for filtering by zodiac
//   - status: index for filtering drafts/published
// Reference: DDIA Ch. 2 - Denormalization (author_zodiac) reduces query complexity
type Post struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"-"` // Hidden for anonymity
	AuthorZodiac string            `bson:"author_zodiac" json:"author_zodiac"` // Denormalized
	
	Title       string   `bson:"title" json:"title"`
	Content     string   `bson:"content" json:"content"`
	MoodTags    []string `bson:"mood_tags" json:"mood_tags"` // Embedded array
	
	Status      PostStatus `bson:"status" json:"status"`
	LikesCount  int        `bson:"likes_count" json:"likes_count"`
	CommentsCount int      `bson:"comments_count" json:"comments_count"`
	
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Like represents a post like (for preventing double-likes)
// Indexes:
//   - {post_id: 1, user_id: 1}: unique compound index prevents double-like
type Like struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PostID    primitive.ObjectID `bson:"post_id" json:"post_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// Comment represents a post comment
type Comment struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	PostID    primitive.ObjectID  `bson:"post_id" json:"post_id"`
	UserID    primitive.ObjectID  `bson:"user_id" json:"user_id"`
	Username  string              `bson:"username" json:"username"` // Denormalized
	Content   string              `bson:"content" json:"content"`
	ParentID  *primitive.ObjectID `bson:"parent_id,omitempty" json:"parent_id,omitempty"` // For nested comments
	CreatedAt time.Time           `bson:"created_at" json:"created_at"`
}

// PublishPostRequest represents publish post request
type PublishPostRequest struct {
	Title    string   `json:"title" validate:"required"`
	Content  string   `json:"content" validate:"required"`
	MoodTags []string `json:"mood_tags"`
}

// AddCommentRequest represents add comment request
type AddCommentRequest struct {
	Content  string `json:"content" validate:"required"`
	ParentID string `json:"parent_id,omitempty"` // Optional for nested comments
}

// GetFeedQuery represents feed query parameters
type GetFeedQuery struct {
	Cursor     string `query:"cursor"`
	Limit      int    `query:"limit"`
	ZodiacSign string `query:"zodiac"`
	Mood       string `query:"mood"`
	SortBy     string `query:"sort"` // latest, most_liked
}
