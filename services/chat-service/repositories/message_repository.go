package repositories

import (
	"context"
	"time"

	"zodiac-ai-backend/services/chat-service/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MessageRepository handles message data access
type MessageRepository struct {
	collection *mongo.Collection
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{
		collection: db.Collection("messages"),
	}
}

// Create creates a new message
func (r *MessageRepository) Create(ctx context.Context, message *models.Message) error {
	message.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		return err
	}

	message.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindBySessionID finds messages by session ID with cursor-based pagination
// Reference: CLRS Ch. 12 - Cursor-based pagination with O(log n) complexity
func (r *MessageRepository) FindBySessionID(ctx context.Context, sessionID primitive.ObjectID, cursor string, limit int) ([]*models.Message, string, error) {
	filter := bson.M{"session_id": sessionID}

	// If cursor provided, filter messages before cursor
	if cursor != "" {
		cursorID, err := primitive.ObjectIDFromHex(cursor)
		if err == nil {
			filter["_id"] = bson.M{"$lt": cursorID}
		}
	}

	opts := options.Find().
		SetSort(bson.M{"_id": -1}). // Sort by _id descending (newest first)
		SetLimit(int64(limit + 1))   // Fetch one extra to check if there's more

	cur, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	defer cur.Close(ctx)

	var messages []*models.Message
	if err := cur.All(ctx, &messages); err != nil {
		return nil, "", err
	}

	// Check if there are more messages
	var nextCursor string
	if len(messages) > limit {
		// Remove the extra message
		messages = messages[:limit]
		// Set next cursor to the last message ID
		nextCursor = messages[len(messages)-1].ID.Hex()
	}

	return messages, nextCursor, nil
}

// GetAllBySessionID gets all messages for a session (for insight generation)
func (r *MessageRepository) GetAllBySessionID(ctx context.Context, sessionID primitive.ObjectID) ([]*models.Message, error) {
	opts := options.Find().SetSort(bson.M{"created_at": 1}) // Chronological order

	cursor, err := r.collection.Find(ctx, bson.M{"session_id": sessionID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// CountBySessionID counts messages in a session
func (r *MessageRepository) CountBySessionID(ctx context.Context, sessionID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"session_id": sessionID})
}
