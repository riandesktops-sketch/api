package services

import (
	"context"
	"errors"

	"zodiac-ai-backend/services/auth-service/models"
	"zodiac-ai-backend/services/auth-service/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrAlreadyFriends    = errors.New("already friends")
	ErrRequestNotFound   = errors.New("friend request not found")
	ErrUnauthorized      = errors.New("unauthorized action")
)

// FriendshipService handles friendship business logic
type FriendshipService struct {
	friendshipRepo *repositories.FriendshipRepository
	userRepo       *repositories.UserRepository
}

// NewFriendshipService creates a new friendship service
func NewFriendshipService(
	friendshipRepo *repositories.FriendshipRepository,
	userRepo *repositories.UserRepository,
) *FriendshipService {
	return &FriendshipService{
		friendshipRepo: friendshipRepo,
		userRepo:       userRepo,
	}
}

// SendFriendRequest sends a friend request
func (s *FriendshipService) SendFriendRequest(ctx context.Context, senderID, targetID string) error {
	senderObjID, err := primitive.ObjectIDFromHex(senderID)
	if err != nil {
		return err
	}

	targetObjID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return err
	}

	// Check if already friends
	areFriends, err := s.friendshipRepo.CheckFriendship(ctx, senderObjID, targetObjID)
	if err != nil {
		return err
	}
	if areFriends {
		return ErrAlreadyFriends
	}

	// Ensure both users have friendship documents
	_, err = s.friendshipRepo.GetOrCreateFriendship(ctx, senderObjID)
	if err != nil {
		return err
	}
	_, err = s.friendshipRepo.GetOrCreateFriendship(ctx, targetObjID)
	if err != nil {
		return err
	}

	// Create friend request
	_, err = s.friendshipRepo.CreateFriendRequest(ctx, senderObjID, targetObjID)
	if err != nil {
		return err
	}

	// Update friendship documents
	if err := s.friendshipRepo.AddPendingSent(ctx, senderObjID, targetObjID); err != nil {
		return err
	}
	if err := s.friendshipRepo.AddPendingReceived(ctx, targetObjID, senderObjID); err != nil {
		return err
	}

	return nil
}

// AcceptFriendRequest accepts a friend request
func (s *FriendshipService) AcceptFriendRequest(ctx context.Context, requestID, userID string) error {
	reqObjID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Find request
	request, err := s.friendshipRepo.FindRequestByID(ctx, reqObjID)
	if err != nil {
		return ErrRequestNotFound
	}

	// Verify user is the receiver
	if request.ReceiverID != userObjID {
		return ErrUnauthorized
	}

	// Accept friendship (transaction)
	if err := s.friendshipRepo.AcceptFriendship(ctx, userObjID, request.SenderID); err != nil {
		return err
	}

	// Update request status
	if err := s.friendshipRepo.UpdateRequestStatus(ctx, reqObjID, models.StatusAccepted); err != nil {
		return err
	}

	// Increment friends count for both users (atomic)
	if err := s.userRepo.IncrementFriendsCount(ctx, userObjID, 1); err != nil {
		return err
	}
	if err := s.userRepo.IncrementFriendsCount(ctx, request.SenderID, 1); err != nil {
		return err
	}

	return nil
}

// RejectFriendRequest rejects a friend request
func (s *FriendshipService) RejectFriendRequest(ctx context.Context, requestID, userID string) error {
	reqObjID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Find request
	request, err := s.friendshipRepo.FindRequestByID(ctx, reqObjID)
	if err != nil {
		return ErrRequestNotFound
	}

	// Verify user is the receiver
	if request.ReceiverID != userObjID {
		return ErrUnauthorized
	}

	// Reject friendship
	if err := s.friendshipRepo.RejectFriendship(ctx, userObjID, request.SenderID); err != nil {
		return err
	}

	// Update request status
	if err := s.friendshipRepo.UpdateRequestStatus(ctx, reqObjID, models.StatusRejected); err != nil {
		return err
	}

	return nil
}

// GetFriends gets list of friends
func (s *FriendshipService) GetFriends(ctx context.Context, userID string) ([]primitive.ObjectID, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return s.friendshipRepo.GetFriends(ctx, userObjID)
}

// CheckFriendshipStatus checks friendship status between two users
// O(1) complexity with compound index
func (s *FriendshipService) CheckFriendshipStatus(ctx context.Context, userID, targetID string) (string, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", err
	}

	targetObjID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		return "", err
	}

	// Check if friends
	areFriends, err := s.friendshipRepo.CheckFriendship(ctx, userObjID, targetObjID)
	if err != nil {
		return "", err
	}

	if areFriends {
		return "ARE_FRIENDS", nil
	}

	// Check pending status
	friendship, err := s.friendshipRepo.GetOrCreateFriendship(ctx, userObjID)
	if err != nil {
		return "", err
	}

	// Check if pending sent
	for _, id := range friendship.PendingSent {
		if id == targetObjID {
			return "PENDING", nil
		}
	}

	// Check if pending received
	for _, id := range friendship.PendingReceived {
		if id == targetObjID {
			return "PENDING", nil
		}
	}

	return "NOT_FRIENDS", nil
}
