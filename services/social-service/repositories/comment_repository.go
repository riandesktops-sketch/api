package repositories

import (
	"context"
	"time"

	"zodiac-ai-backend/services/social-service/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CommentRepository handles comment data access
type CommentRepository struct {
	collection *mongo.Collection
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *mongo.Database) *CommentRepository {
	return &CommentRepository{
		collection: db.Collection("comments"),
	}
}

// Create creates a new comment
func (r *CommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	comment.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, comment)
	if err != nil {
		return err
	}

	comment.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByPostID finds comments by post ID
func (r *CommentRepository) FindByPostID(ctx context.Context, postID primitive.ObjectID) ([]*models.Comment, error) {
	opts := options.Find().SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{"post_id": postID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []*models.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// CountByPostID counts comments for a post
func (r *CommentRepository) CountByPostID(ctx context.Context, postID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"post_id": postID})
}
