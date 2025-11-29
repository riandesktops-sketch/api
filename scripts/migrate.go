package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"zodiac-ai-backend/pkg/config"
	"zodiac-ai-backend/pkg/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	log.Println("🚀 Starting MongoDB migration...")

	// Load config
	cfg := config.LoadConfig()

	// Connect to MongoDB
	_, err := database.Connect(database.MongoConfig{
		URI:      cfg.MongoURI,
		Database: cfg.MongoDatabase,
	})
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	db := database.GetDatabase(cfg.MongoDatabase)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run migrations
	if err := migrateUsers(ctx, db); err != nil {
		log.Fatalf("Failed to migrate users: %v", err)
	}

	if err := migrateFriendships(ctx, db); err != nil {
		log.Fatalf("Failed to migrate friendships: %v", err)
	}

	if err := migrateRefreshTokens(ctx, db); err != nil {
		log.Fatalf("Failed to migrate refresh tokens: %v", err)
	}

	if err := migrateMessages(ctx, db); err != nil {
		log.Fatalf("Failed to migrate messages: %v", err)
	}

	if err := migrateRooms(ctx, db); err != nil {
		log.Fatalf("Failed to migrate rooms: %v", err)
	}

	if err := migrateRoomMessages(ctx, db); err != nil {
		log.Fatalf("Failed to migrate room messages: %v", err)
	}

	if err := migratePosts(ctx, db); err != nil {
		log.Fatalf("Failed to migrate posts: %v", err)
	}

	if err := migrateLikes(ctx, db); err != nil {
		log.Fatalf("Failed to migrate likes: %v", err)
	}

	if err := migrateComments(ctx, db); err != nil {
		log.Fatalf("Failed to migrate comments: %v", err)
	}

	log.Println("✅ Migration completed successfully!")
}

// migrateUsers creates indexes for users collection
func migrateUsers(ctx context.Context, db *mongo.Database) error {
	log.Println("Migrating users collection...")
	coll := db.Collection("users")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "zodiac_sign", Value: 1}},
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create users indexes: %w", err)
	}

	log.Println("✅ Users collection migrated")
	return nil
}

// migrateFriendships creates indexes for friendships collection
func migrateFriendships(ctx context.Context, db *mongo.Database) error {
	log.Println("Migrating friendships collection...")
	coll := db.Collection("friendships")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			// Compound index for O(1) friendship lookup
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "friend_ids", Value: 1},
			},
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create friendships indexes: %w", err)
	}

	log.Println("✅ Friendships collection migrated")
	return nil
}

// migrateRefreshTokens creates indexes for refresh_tokens collection
func migrateRefreshTokens(ctx context.Context, db *mongo.Database) error {
	log.Println("Migrating refresh_tokens collection...")
	coll := db.Collection("refresh_tokens")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			// TTL index: auto-delete after 30 days (2592000 seconds)
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(2592000),
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create refresh_tokens indexes: %w", err)
	}

	log.Println("✅ Refresh tokens collection migrated (TTL: 30 days)")
	return nil
}

// migrateMessages creates indexes for messages collection
func migrateMessages(ctx context.Context, db *mongo.Database) error {
	log.Println("Migrating messages collection...")
	coll := db.Collection("messages")

	indexes := []mongo.IndexModel{
		{
			// TTL index: auto-delete after 48 hours (172800 seconds)
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(172800),
		},
		{
			Keys: bson.D{{Key: "session_id", Value: 1}},
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create messages indexes: %w", err)
	}

	log.Println("✅ Messages collection migrated (TTL: 48 hours)")
	return nil
}

// migrateRooms creates indexes for rooms collection
func migrateRooms(ctx context.Context, db *mongo.Database) error {
	log.Println("Migrating rooms collection...")
	coll := db.Collection("rooms")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "topic", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "zodiac_filter", Value: 1}},
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create rooms indexes: %w", err)
	}

	log.Println("✅ Rooms collection migrated")
	return nil
}

// migrateRoomMessages creates indexes for room_messages collection
func migrateRoomMessages(ctx context.Context, db *mongo.Database) error {
	log.Println("Migrating room_messages collection...")
	coll := db.Collection("room_messages")

	indexes := []mongo.IndexModel{
		{
			// TTL index: auto-delete after 24 hours (86400 seconds)
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(86400),
		},
		{
			Keys: bson.D{{Key: "room_id", Value: 1}},
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create room_messages indexes: %w", err)
	}

	log.Println("✅ Room messages collection migrated (TTL: 24 hours)")
	return nil
}

// migratePosts creates indexes for posts collection
func migratePosts(ctx context.Context, db *mongo.Database) error {
	log.Println("Migrating posts collection...")
	coll := db.Collection("posts")

	indexes := []mongo.IndexModel{
		{
			// Compound index for feed sorting (latest and most liked)
			Keys: bson.D{
				{Key: "created_at", Value: -1},
				{Key: "likes_count", Value: -1},
			},
		},
		{
			Keys: bson.D{{Key: "author_zodiac", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create posts indexes: %w", err)
	}

	log.Println("✅ Posts collection migrated (NO TTL - permanent storage)")
	return nil
}

// migrateLikes creates indexes for likes collection
func migrateLikes(ctx context.Context, db *mongo.Database) error {
	log.Println("Migrating likes collection...")
	coll := db.Collection("likes")

	indexes := []mongo.IndexModel{
		{
			// Unique compound index to prevent double-like
			Keys: bson.D{
				{Key: "post_id", Value: 1},
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create likes indexes: %w", err)
	}

	log.Println("✅ Likes collection migrated")
	return nil
}

// migrateComments creates indexes for comments collection
func migrateComments(ctx context.Context, db *mongo.Database) error {
	log.Println("Migrating comments collection...")
	coll := db.Collection("comments")

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "post_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "parent_id", Value: 1}},
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create comments indexes: %w", err)
	}

	log.Println("✅ Comments collection migrated")
	return nil
}
