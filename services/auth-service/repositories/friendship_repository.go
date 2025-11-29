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
	ErrFriendshipNotFound = errors.New("friendship not found")
)

// FriendshipRepository handles friendship data access
// Reference: CLRS Ch. 22 - Graph representation using adjacency list
type FriendshipRepository struct {
	collection        *mongo.Collection
	requestCollection *mongo.Collection
}

// NewFriendshipRepository creates a new friendship repository
func NewFriendshipRepository(db *mongo.Database) *FriendshipRepository {
	return &FriendshipRepository{
		collection:        db.Collection("friendships"),
		requestCollection: db.Collection("friend_requests"),
	}
}

// GetOrCreateFriendship gets or creates friendship document for user
func (r *FriendshipRepository) GetOrCreateFriendship(ctx context.Context, userID primitive.ObjectID) (*models.Friendship, error) {
	var friendship models.Friendship
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&friendship)
	
	if err == mongo.ErrNoDocuments {
		// Create new friendship document
		friendship = models.Friendship{
			UserID:          userID,
			FriendIDs:       []primitive.ObjectID{},
			PendingSent:     []primitive.ObjectID{},
			PendingReceived: []primitive.ObjectID{},
			UpdatedAt:       time.Now(),
		}
		
		result, err := r.collection.InsertOne(ctx, &friendship)
		if err != nil {
			return nil, err
		}
		friendship.ID = result.InsertedID.(primitive.ObjectID)
		return &friendship, nil
	}
	
	if err != nil {
		return nil, err
	}
	
	return &friendship, nil
}

// CheckFriendship checks if two users are friends
// O(1) complexity using $in operator with compound index
// Reference: CLRS - Array membership check with index
func (r *FriendshipRepository) CheckFriendship(ctx context.Context, userID, friendID primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"user_id":    userID,
		"friend_ids": bson.M{"$in": []primitive.ObjectID{friendID}},
	})
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// AddPendingSent adds a pending sent request
func (r *FriendshipRepository) AddPendingSent(ctx context.Context, userID, targetID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{
			"$addToSet": bson.M{"pending_sent": targetID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// AddPendingReceived adds a pending received request
func (r *FriendshipRepository) AddPendingReceived(ctx context.Context, userID, senderID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{
			"$addToSet": bson.M{"pending_received": senderID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// AcceptFriendship accepts a friend request (bidirectional)
// Uses MongoDB session for transaction to prevent race conditions
// Reference: DDIA Ch. 7 - Transactions for atomicity
func (r *FriendshipRepository) AcceptFriendship(ctx context.Context, userID, friendID primitive.ObjectID) error {
	// Start session for transaction
	session, err := r.collection.Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Execute transaction
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Update user's friendship
		_, err := r.collection.UpdateOne(
			sessCtx,
			bson.M{"user_id": userID},
			bson.M{
				"$addToSet": bson.M{"friend_ids": friendID},
				"$pull":     bson.M{"pending_received": friendID},
				"$set":      bson.M{"updated_at": time.Now()},
			},
		)
		if err != nil {
			return nil, err
		}

		// Update friend's friendship (bidirectional)
		_, err = r.collection.UpdateOne(
			sessCtx,
			bson.M{"user_id": friendID},
			bson.M{
				"$addToSet": bson.M{"friend_ids": userID},
				"$pull":     bson.M{"pending_sent": userID},
				"$set":      bson.M{"updated_at": time.Now()},
			},
		)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

// RejectFriendship rejects a friend request
func (r *FriendshipRepository) RejectFriendship(ctx context.Context, userID, friendID primitive.ObjectID) error {
	session, err := r.collection.Database().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Remove from user's pending received
		_, err := r.collection.UpdateOne(
			sessCtx,
			bson.M{"user_id": userID},
			bson.M{
				"$pull": bson.M{"pending_received": friendID},
				"$set":  bson.M{"updated_at": time.Now()},
			},
		)
		if err != nil {
			return nil, err
		}

		// Remove from friend's pending sent
		_, err = r.collection.UpdateOne(
			sessCtx,
			bson.M{"user_id": friendID},
			bson.M{
				"$pull": bson.M{"pending_sent": userID},
				"$set":  bson.M{"updated_at": time.Now()},
			},
		)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

// GetFriends gets list of friends for a user
func (r *FriendshipRepository) GetFriends(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	friendship, err := r.GetOrCreateFriendship(ctx, userID)
	if err != nil {
		return nil, err
	}
	return friendship.FriendIDs, nil
}

// CreateFriendRequest creates a friend request
func (r *FriendshipRepository) CreateFriendRequest(ctx context.Context, senderID, receiverID primitive.ObjectID) (*models.FriendRequest, error) {
	request := &models.FriendRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Status:     models.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	result, err := r.requestCollection.InsertOne(ctx, request)
	if err != nil {
		return nil, err
	}

	request.ID = result.InsertedID.(primitive.ObjectID)
	return request, nil
}

// UpdateRequestStatus updates friend request status
func (r *FriendshipRepository) UpdateRequestStatus(ctx context.Context, requestID primitive.ObjectID, status models.FriendshipStatus) error {
	_, err := r.requestCollection.UpdateOne(
		ctx,
		bson.M{"_id": requestID},
		bson.M{
			"$set": bson.M{
				"status":     status,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}

// FindRequestByID finds a friend request by ID
func (r *FriendshipRepository) FindRequestByID(ctx context.Context, requestID primitive.ObjectID) (*models.FriendRequest, error) {
	var request models.FriendRequest
	err := r.requestCollection.FindOne(ctx, bson.M{"_id": requestID}).Decode(&request)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("friend request not found")
		}
		return nil, err
	}
	return &request, nil
}
