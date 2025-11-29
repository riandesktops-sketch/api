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

// ChatSessionRepository handles chat session data access
type ChatSessionRepository struct {
	collection *mongo.Collection
}

// NewChatSessionRepository creates a new chat session repository
func NewChatSessionRepository(db *mongo.Database) *ChatSessionRepository {
	return &ChatSessionRepository{
		collection: db.Collection("chat_sessions"),
	}
}

// Create creates a new chat session
func (r *ChatSessionRepository) Create(ctx context.Context, userID primitive.ObjectID, title string) (*models.ChatSession, error) {
	session := &models.ChatSession{
		UserID:    userID,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := r.collection.InsertOne(ctx, session)
	if err != nil {
		return nil, err
	}

	session.ID = result.InsertedID.(primitive.ObjectID)
	return session, nil
}

// FindByID finds a chat session by ID
func (r *ChatSessionRepository) FindByID(ctx context.Context, sessionID primitive.ObjectID) (*models.ChatSession, error) {
	var session models.ChatSession
	err := r.collection.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindByUserID finds all chat sessions for a user
func (r *ChatSessionRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.ChatSession, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []*models.ChatSession
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// Update updates a chat session
func (r *ChatSessionRepository) Update(ctx context.Context, sessionID primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": sessionID},
		bson.M{"$set": update},
	)
	return err
}
