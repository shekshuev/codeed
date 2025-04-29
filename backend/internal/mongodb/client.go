package mongodb

import (
	"context"
	"sync"

	"github.com/shekshuev/codeed/backend/internal/config"
	"github.com/shekshuev/codeed/backend/internal/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Mongo wraps the MongoDB client instance used across the application.
// It is initialized only once using a singleton pattern.
type Mongo struct {
	Client *mongo.Client
}

var (
	instance *Mongo
	once     sync.Once
)

// NewMongoClient initializes and returns a singleton Mongo client wrapper.
// It uses the URI from the provided config to connect to the database and performs a ping to ensure connectivity.
// If already initialized, it returns the existing instance.
func NewMongoClient(ctx context.Context, cfg *config.Config) *Mongo {
	once.Do(func() {
		logger := logger.NewLogger().Log
		clientOpts := options.Client().ApplyURI(cfg.MongoURI)
		client, err := mongo.Connect(ctx, clientOpts)
		if err != nil {
			logger.Fatal("Mongo connect error:", zap.Error(err))
		}
		if err := client.Ping(ctx, nil); err != nil {
			logger.Fatal("Mongo ping error:", zap.Error(err))
		}
		instance = &Mongo{Client: client}
	})
	return instance
}
