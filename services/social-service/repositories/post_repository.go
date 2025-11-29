package repositories

import (
	"context"
	"errors"
	"time"

	"zodiac-ai-backend/services/social-service/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrPostNotFound = errors.New("post not found")
)

// PostRepository handles post data access
type PostRepository struct {
	collection     *mongo.Collection
	likeCollection *mongo.Collection
}

// NewPostRepository creates a new post repository
func NewPostRepository(db *mongo.Database) *PostRepository {
	return &PostRepository{
		collection:     db.Collection("posts"),
		likeCollection: db.Collection("likes"),
	}
}

// Create creates a new post
func (r *PostRepository) Create(ctx context.Context, post *models.Post) error {
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	post.LikesCount = 0
	post.CommentsCount = 0

	result, err := r.collection.InsertOne(ctx, post)
	if err != nil {
		return err
	}

	post.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID finds a post by ID
func (r *PostRepository) FindByID(ctx context.Context, postID primitive.ObjectID) (*models.Post, error) {
	var post models.Post
	err := r.collection.FindOne(ctx, bson.M{"_id": postID}).Decode(&post)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return &post, nil
}

// GetFeed gets posts with cursor-based pagination and filters
// Reference: CLRS Ch. 12 - Cursor pagination with O(log n) complexity
func (r *PostRepository) GetFeed(ctx context.Context, query *models.GetFeedQuery) ([]*models.Post, string, error) {
	filter := bson.M{"status": models.StatusPublished}

	// Apply filters
	if query.ZodiacSign != "" {
		filter["author_zodiac"] = query.ZodiacSign
	}
	if query.Mood != "" {
		filter["mood_tags"] = query.Mood
	}

	// Cursor pagination
	if query.Cursor != "" {
		cursorID, err := primitive.ObjectIDFromHex(query.Cursor)
		if err == nil {
			filter["_id"] = bson.M{"$lt": cursorID}
		}
	}

	// Sorting
	sort := bson.M{"_id": -1} // Default: latest
	if query.SortBy == "most_liked" {
		sort = bson.M{"likes_count": -1, "_id": -1}
	}

	opts := options.Find().
		SetSort(sort).
		SetLimit(int64(query.Limit + 1))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	defer cursor.Close(ctx)

	var posts []*models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, "", err
	}

	// Check if there are more posts
	var nextCursor string
	if len(posts) > query.Limit {
		posts = posts[:query.Limit]
		nextCursor = posts[len(posts)-1].ID.Hex()
	}

	return posts, nextCursor, nil
}

// LikePost likes a post (atomic increment)
// Reference: DDIA Ch. 9 - Atomic operations prevent race conditions
func (r *PostRepository) LikePost(ctx context.Context, postID, userID primitive.ObjectID) error {
	// Try to insert like record (unique index prevents double-like)
	like := &models.Like{
		PostID:    postID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	_, err := r.likeCollection.InsertOne(ctx, like)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("already liked")
		}
		return err
	}

	// Atomic increment likes count
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": postID},
		bson.M{"$inc": bson.M{"likes_count": 1}},
	)

	return err
}

// UnlikePost unlikes a post (atomic decrement)
func (r *PostRepository) UnlikePost(ctx context.Context, postID, userID primitive.ObjectID) error {
	// Delete like record
	result, err := r.likeCollection.DeleteOne(ctx, bson.M{
		"post_id": postID,
		"user_id": userID,
	})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("like not found")
	}

	// Atomic decrement likes count
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": postID},
		bson.M{"$inc": bson.M{"likes_count": -1}},
	)

	return err
}

// IncrementCommentsCount increments post's comments count
func (r *PostRepository) IncrementCommentsCount(ctx context.Context, postID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": postID},
		bson.M{"$inc": bson.M{"comments_count": 1}},
	)
	return err
}

// CheckUserLiked checks if user has liked a post
func (r *PostRepository) CheckUserLiked(ctx context.Context, postID, userID primitive.ObjectID) (bool, error) {
	count, err := r.likeCollection.CountDocuments(ctx, bson.M{
		"post_id": postID,
		"user_id": userID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
