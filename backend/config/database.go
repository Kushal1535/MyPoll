package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectMongoDB() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	// Use MongoDB Atlas URI when configured.
	// Otherwise, use local MongoDB for development.
	mongoURI := strings.TrimSpace(os.Getenv("MONGODB_URI"))

	if mongoURI == "" {
		mongoURI = "mongodb://127.0.0.1:27017"
	}

	client, err := mongo.Connect(
		options.Client().ApplyURI(mongoURI),
	)
	if err != nil {
		return nil, fmt.Errorf("MongoDB connection failed: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("MongoDB ping failed: %w", err)
	}

	return client, nil
}
