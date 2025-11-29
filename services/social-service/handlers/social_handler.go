package handlers

import (
	"zodiac-ai-backend/pkg/middleware"
	"zodiac-ai-backend/pkg/response"
	"zodiac-ai-backend/services/social-service/models"
	"zodiac-ai-backend/services/social-service/services"

	"github.com/gofiber/fiber/v2"
)

// SocialHandler handles social feed HTTP requests
type SocialHandler struct {
	socialService *services.SocialService
}

// NewSocialHandler creates a new social handler
func NewSocialHandler(socialService *services.SocialService) *SocialHandler {
	return &SocialHandler{
		socialService: socialService,
	}
}

// PublishPost publishes a new post
// POST /posts
func (h *SocialHandler) PublishPost(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	zodiacSign := middleware.GetZodiacSign(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req models.PublishPostRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	post, err := h.socialService.PublishPost(c.Context(), userID, zodiacSign, &req)
	if err != nil {
		return response.InternalServerError(c, "Failed to publish post")
	}

	return response.Created(c, "Post published successfully", post)
}

// GetFeed gets social feed
// GET /posts
func (h *SocialHandler) GetFeed(c *fiber.Ctx) error {
	query := &models.GetFeedQuery{
		Cursor:     c.Query("cursor", ""),
		Limit:      c.QueryInt("limit", 20),
		ZodiacSign: c.Query("zodiac", ""),
		Mood:       c.Query("mood", ""),
		SortBy:     c.Query("sort", "latest"),
	}

	posts, nextCursor, err := h.socialService.GetFeed(c.Context(), query)
	if err != nil {
		return response.InternalServerError(c, "Failed to get feed")
	}

	meta := &response.MetaData{
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
		Limit:      query.Limit,
	}

	return response.SuccessWithMeta(c, "Feed retrieved successfully", posts, meta)
}

// GetPost gets a single post
// GET /posts/:id
func (h *SocialHandler) GetPost(c *fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID required", nil)
	}

	post, err := h.socialService.GetPost(c.Context(), postID)
	if err != nil {
		return response.NotFound(c, "Post not found")
	}

	return response.Success(c, "Post retrieved successfully", post)
}

// LikePost likes a post
// POST /posts/:id/like
func (h *SocialHandler) LikePost(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID required", nil)
	}

	err := h.socialService.LikePost(c.Context(), postID, userID)
	if err != nil {
		if err.Error() == "already liked" {
			return response.Conflict(c, "Post already liked")
		}
		return response.InternalServerError(c, "Failed to like post")
	}

	return response.Success(c, "Post liked successfully", nil)
}

// UnlikePost unlikes a post
// DELETE /posts/:id/like
func (h *SocialHandler) UnlikePost(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID required", nil)
	}

	err := h.socialService.UnlikePost(c.Context(), postID, userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to unlike post")
	}

	return response.Success(c, "Post unliked successfully", nil)
}

// AddComment adds a comment to a post
// POST /posts/:id/comments
func (h *SocialHandler) AddComment(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID required", nil)
	}

	var req models.AddCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	// Get username from context (you might want to fetch from user service)
	username := c.Locals("username")
	if username == nil {
		username = "Anonymous"
	}

	comment, err := h.socialService.AddComment(c.Context(), postID, userID, username.(string), &req)
	if err != nil {
		return response.InternalServerError(c, "Failed to add comment")
	}

	return response.Created(c, "Comment added successfully", comment)
}

// GetComments gets comments for a post
// GET /posts/:id/comments
func (h *SocialHandler) GetComments(c *fiber.Ctx) error {
	postID := c.Params("id")
	if postID == "" {
		return response.BadRequest(c, "Post ID required", nil)
	}

	comments, err := h.socialService.GetComments(c.Context(), postID)
	if err != nil {
		return response.InternalServerError(c, "Failed to get comments")
	}

	return response.Success(c, "Comments retrieved successfully", comments)
}
