package testutils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetupMongo spins up a temporary MongoDB container for integration testing.
// It returns a connected mongo.Database and a cleanup function.
func SetupMongo(t *testing.T) (*mongo.Database, func()) {
	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:6")
	assert.NoError(t, err)

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	uri, err := container.ConnectionString(ctx)
	assert.NoError(t, err)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	assert.NoError(t, err)

	db := client.Database("testdb")

	return db, func() {
		_ = client.Disconnect(ctx)
	}
}
