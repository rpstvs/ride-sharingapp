package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	TripsCollection     = "trips"
	RidesFareCollection = "ride_fares"
)

type MongoConfig struct {
	URI      string
	Database string
}

func NewMongoConfig() *MongoConfig {
	return &MongoConfig{
		URI:      os.Getenv("MONGODB_URI"),
		Database: "ride-sharing",
	}
}

func NewMongoClient(ctx context.Context, cfg *MongoConfig) (*mongo.Client, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("mongo db uri is required")
	}

	if cfg.Database == "" {
		return nil, fmt.Errorf("mongo db database is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))

	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetDatabase(client *mongo.Client, cfg *MongoConfig) *mongo.Database {
	return client.Database(cfg.Database)
}
