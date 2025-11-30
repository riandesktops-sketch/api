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

// RoomRepository handles room data access
type RoomRepository struct {
	collection        *mongo.Collection
	messageCollection *mongo.Collection
}

// NewRoomRepository creates a new room repository
func NewRoomRepository(db *mongo.Database) *RoomRepository {
	return &RoomRepository{
		collection:        db.Collection("rooms"),
		messageCollection: db.Collection("room_messages"),
	}
}

// Create creates a new room
func (r *RoomRepository) Create(ctx context.Context, room *models.Room) error {
	room.CreatedAt = time.Now()
	room.UpdatedAt = time.Now()
	room.MemberCount = 0

	result, err := r.collection.InsertOne(ctx, room)
	if err != nil {
		return err
	}

	room.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID finds a room by ID
func (r *RoomRepository) FindByID(ctx context.Context, roomID primitive.ObjectID) (*models.Room, error) {
	var room models.Room
	err := r.collection.FindOne(ctx, bson.M{"_id": roomID}).Decode(&room)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

// GetRooms gets all rooms with filters
func (r *RoomRepository) GetRooms(ctx context.Context, topic, zodiacFilter string, limit int) ([]*models.Room, error) {
	filter := bson.M{}

	if topic != "" {
		filter["topic"] = topic
	}
	if zodiacFilter != "" {
		filter["zodiac_filter"] = zodiacFilter
	}

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []*models.Room
	if err := cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}

	return rooms, nil
}

// SaveMessage saves a room message
func (r *RoomRepository) SaveMessage(ctx context.Context, message *models.RoomMessage) error {
	message.CreatedAt = time.Now()

	_, err := r.messageCollection.InsertOne(ctx, message)
	return err
}

// UpdateLastMessage updates room's last message
func (r *RoomRepository) UpdateLastMessage(ctx context.Context, roomID primitive.ObjectID, preview *models.RoomMessagePreview) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": roomID},
		bson.M{
			"$set": bson.M{
				"last_message": preview,
				"updated_at":   time.Now(),
			},
		},
	)
	return err
}

// IncrementMemberCount increments room's member count
func (r *RoomRepository) IncrementMemberCount(ctx context.Context, roomID primitive.ObjectID, delta int) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": roomID},
		bson.M{"$inc": bson.M{"member_count": delta}},
	)
	return err
}

// Delete deletes a room and all its messages
func (r *RoomRepository) Delete(ctx context.Context, roomID primitive.ObjectID) error {
	// Delete all messages in the room first
	_, err := r.messageCollection.DeleteMany(ctx, bson.M{"room_id": roomID})
	if err != nil {
		return err
	}

	// Delete the room
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": roomID})
	return err
}

