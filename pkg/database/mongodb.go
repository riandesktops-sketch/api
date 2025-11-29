package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	client *mongo.Client
	once   sync.Once
)

// MongoConfig holds MongoDB connection configuration
type MongoConfig struct {
	URI      string
	Database string
}

// Connect establishes MongoDB connection with connection pooling
// Implements singleton pattern for efficient resource usage
// Reference: DDIA Ch. 5 - Connection pooling optimizes for read/write performance
func Connect(cfg MongoConfig) (*mongo.Client, error) {
	var err error

	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Client options with connection pooling
		clientOptions := options.Client().
			ApplyURI(cfg.URI).
			SetMaxPoolSize(100).                    // Max 100 connections in pool
			SetMinPoolSize(10).                     // Min 10 connections always ready
			SetMaxConnIdleTime(30 * time.Second).   // Close idle connections after 30s
			SetServerSelectionTimeout(5 * time.Second) // Timeout for server selection

		// Create client
		client, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Printf("Failed to create MongoDB client: %v", err)
			return
		}

		// Ping to verify connection
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()

		err = client.Ping(ctx2, readpref.Primary())
		if err != nil {
			log.Printf("Failed to ping MongoDB: %v", err)
			return
		}

		log.Println("✅ MongoDB connected successfully")
	})

	if err != nil {
		return nil, fmt.Errorf("mongodb connection error: %w", err)
	}

	return client, nil
}

// GetClient returns the singleton MongoDB client
func GetClient() *mongo.Client {
	if client == nil {
		log.Fatal("MongoDB client not initialized. Call Connect() first.")
	}
	return client
}

// GetDatabase returns a database instance
func GetDatabase(name string) *mongo.Database {
	return GetClient().Database(name)
}

// GetCollection returns a collection instance
func GetCollection(database, collection string) *mongo.Collection {
	return GetDatabase(database).Collection(collection)
}

// Disconnect closes MongoDB connection gracefully
func Disconnect() error {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("failed to disconnect MongoDB: %w", err)
	}

	log.Println("✅ MongoDB disconnected successfully")
	return nil
}

// HealthCheck verifies MongoDB connection is alive
func HealthCheck(ctx context.Context) error {
	if client == nil {
		return fmt.Errorf("mongodb client not initialized")
	}

	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		return fmt.Errorf("mongodb health check failed: %w", err)
	}

	return nil
}
