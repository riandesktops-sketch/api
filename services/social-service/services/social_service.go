package services

import (
	"context"

	"zodiac-ai-backend/services/social-service/models"
	"zodiac-ai-backend/services/social-service/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SocialService handles social feed business logic
type SocialService struct {
	postRepo    *repositories.PostRepository
	commentRepo *repositories.CommentRepository
}

// NewSocialService creates a new social service
func NewSocialService(
	postRepo *repositories.PostRepository,
	commentRepo *repositories.CommentRepository,
) *SocialService {
	return &SocialService{
		postRepo:    postRepo,
		commentRepo: commentRepo,
	}
}

// PublishPost publishes a new post
func (s *SocialService) PublishPost(ctx context.Context, userID, zodiacSign string, req *models.PublishPostRequest) (*models.Post, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	post := &models.Post{
		UserID:       userObjID,
		AuthorZodiac: zodiacSign, // Denormalized for filtering
		Title:        req.Title,
		Content:      req.Content,
		MoodTags:     req.MoodTags,
		Status:       models.StatusPublished,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// GetFeed gets social feed with filters and pagination
func (s *SocialService) GetFeed(ctx context.Context, query *models.GetFeedQuery) ([]*models.Post, string, error) {
	// Set default limit
	if query.Limit <= 0 || query.Limit > 50 {
		query.Limit = 20
	}

	return s.postRepo.GetFeed(ctx, query)
}

// GetPost gets a single post by ID
func (s *SocialService) GetPost(ctx context.Context, postID string) (*models.Post, error) {
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, err
	}

	return s.postRepo.FindByID(ctx, postObjID)
}

// LikePost likes a post
func (s *SocialService) LikePost(ctx context.Context, postID, userID string) error {
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	return s.postRepo.LikePost(ctx, postObjID, userObjID)
}

// UnlikePost unlikes a post
func (s *SocialService) UnlikePost(ctx context.Context, postID, userID string) error {
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	return s.postRepo.UnlikePost(ctx, postObjID, userObjID)
}

// AddComment adds a comment to a post
func (s *SocialService) AddComment(ctx context.Context, postID, userID, username string, req *models.AddCommentRequest) (*models.Comment, error) {
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// Verify post exists
	_, err = s.postRepo.FindByID(ctx, postObjID)
	if err != nil {
		return nil, err
	}

	comment := &models.Comment{
		PostID:   postObjID,
		UserID:   userObjID,
		Username: username, // Denormalized for display
		Content:  req.Content,
	}

	// Handle parent comment (nested replies)
	if req.ParentID != "" {
		parentObjID, err := primitive.ObjectIDFromHex(req.ParentID)
		if err == nil {
			comment.ParentID = &parentObjID
		}
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	// Increment post's comments count (atomic)
	if err := s.postRepo.IncrementCommentsCount(ctx, postObjID); err != nil {
		return nil, err
	}

	return comment, nil
}

// GetComments gets comments for a post
func (s *SocialService) GetComments(ctx context.Context, postID string) ([]*models.Comment, error) {
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, err
	}

	return s.commentRepo.FindByPostID(ctx, postObjID)
}

// CheckUserLiked checks if user has liked a post
func (s *SocialService) CheckUserLiked(ctx context.Context, postID, userID string) (bool, error) {
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return false, err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	return s.postRepo.CheckUserLiked(ctx, postObjID, userObjID)
}
