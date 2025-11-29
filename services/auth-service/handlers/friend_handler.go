package handlers

import (
	"zodiac-ai-backend/pkg/middleware"
	"zodiac-ai-backend/pkg/response"
	"zodiac-ai-backend/services/auth-service/models"
	"zodiac-ai-backend/services/auth-service/services"

	"github.com/gofiber/fiber/v2"
)

// FriendHandler handles friendship HTTP requests
type FriendHandler struct {
	friendshipService *services.FriendshipService
}

// NewFriendHandler creates a new friend handler
func NewFriendHandler(friendshipService *services.FriendshipService) *FriendHandler {
	return &FriendHandler{
		friendshipService: friendshipService,
	}
}

// SendFriendRequest sends a friend request
// POST /friends/requests
func (h *FriendHandler) SendFriendRequest(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	var req models.SendFriendRequestInput
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	err := h.friendshipService.SendFriendRequest(c.Context(), userID, req.TargetUserID)
	if err != nil {
		if err == services.ErrAlreadyFriends {
			return response.Conflict(c, "Already friends")
		}
		return response.InternalServerError(c, "Failed to send friend request")
	}

	return response.Success(c, "Friend request sent successfully", nil)
}

// AcceptRejectRequest accepts or rejects a friend request
// PUT /friends/requests/:id
func (h *FriendHandler) AcceptRejectRequest(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	requestID := c.Params("id")
	if requestID == "" {
		return response.BadRequest(c, "Request ID required", nil)
	}

	var req struct {
		Action string `json:"action" validate:"required,oneof=accept reject"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	var err error
	if req.Action == "accept" {
		err = h.friendshipService.AcceptFriendRequest(c.Context(), requestID, userID)
	} else {
		err = h.friendshipService.RejectFriendRequest(c.Context(), requestID, userID)
	}

	if err != nil {
		if err == services.ErrRequestNotFound {
			return response.NotFound(c, "Friend request not found")
		}
		if err == services.ErrUnauthorized {
			return response.Unauthorized(c, "Unauthorized action")
		}
		return response.InternalServerError(c, "Failed to process friend request")
	}

	message := "Friend request accepted"
	if req.Action == "reject" {
		message = "Friend request rejected"
	}

	return response.Success(c, message, nil)
}

// GetFriends gets list of friends
// GET /friends
func (h *FriendHandler) GetFriends(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	friendIDs, err := h.friendshipService.GetFriends(c.Context(), userID)
	if err != nil {
		return response.InternalServerError(c, "Failed to get friends")
	}

	return response.Success(c, "Friends retrieved successfully", fiber.Map{
		"friend_ids": friendIDs,
		"count":      len(friendIDs),
	})
}

// CheckFriendshipStatus checks friendship status with another user
// GET /friends/status/:user_id
func (h *FriendHandler) CheckFriendshipStatus(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated")
	}

	targetUserID := c.Params("user_id")
	if targetUserID == "" {
		return response.BadRequest(c, "Target user ID required", nil)
	}

	status, err := h.friendshipService.CheckFriendshipStatus(c.Context(), userID, targetUserID)
	if err != nil {
		return response.InternalServerError(c, "Failed to check friendship status")
	}

	return response.Success(c, "Friendship status retrieved", fiber.Map{
		"status": status,
	})
}
