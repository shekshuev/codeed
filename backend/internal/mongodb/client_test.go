package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/shekshuev/codeed/backend/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func TestNewMongoClient_WithTestContainer(t *testing.T) {
	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:6")
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	uri, err := container.ConnectionString(ctx)
	assert.NoError(t, err)

	cfg := &config.Config{
		MongoURI: uri,
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	clientWrapper := NewMongoClient(ctxWithTimeout, cfg)
	assert.NotNil(t, clientWrapper)
	assert.NotNil(t, clientWrapper.Client)

	secondCall := NewMongoClient(ctxWithTimeout, cfg)
	assert.Equal(t, clientWrapper, secondCall)
}
