package repositories

import (
	"context"
	"errors"
	"time"

	"zodiac-ai-backend/services/auth-service/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrTokenNotFound = errors.New("refresh token not found")
)

// RefreshTokenRepository handles refresh token data access
type RefreshTokenRepository struct {
	collection *mongo.Collection
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *mongo.Database) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		collection: db.Collection("refresh_tokens"),
	}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(ctx context.Context, userID primitive.ObjectID, token string) error {
	refreshToken := &models.RefreshToken{
		UserID:    userID,
		Token:     token,
		CreatedAt: time.Now(),
	}

	_, err := r.collection.InsertOne(ctx, refreshToken)
	return err
}

// FindByToken finds a refresh token
func (r *RefreshTokenRepository) FindByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&refreshToken)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}
	return &refreshToken, nil
}

// Delete deletes a refresh token
func (r *RefreshTokenRepository) Delete(ctx context.Context, token string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"token": token})
	return err
}

// DeleteByUserID deletes all refresh tokens for a user
func (r *RefreshTokenRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}
