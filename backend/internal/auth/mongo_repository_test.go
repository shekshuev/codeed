package auth

import (
	"context"
	"testing"
	"time"

	"github.com/shekshuev/codeed/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAuthAttemptRepositoryImpl_Create(t *testing.T) {
	db, disconnect := testutils.SetupMongo(t)
	repoImpl := NewAuthAttemptRepository(db)
	defer disconnect()

	ctx := context.Background()

	t.Run("creates an auth attempt", func(t *testing.T) {
		attempt := AuthAttempt{
			ID:             primitive.NewObjectID(),
			IdentifierUsed: "user1",
			Type:           TypeTelegram,
			Code:           "123456",
			Success:        false,
			AttemptLeft:    3,
			TTL:            AttemptTTL,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
		res, err := repoImpl.Create(ctx, attempt)
		assert.NoError(t, err)
		assert.Equal(t, attempt.IdentifierUsed, res.IdentifierUsed)
	})
}

func TestAuthAttemptRepositoryImpl_GetByID(t *testing.T) {
	db, disconnect := testutils.SetupMongo(t)
	repoImpl := NewAuthAttemptRepository(db)
	defer disconnect()

	ctx := context.Background()
	attempt := AuthAttempt{
		ID:             primitive.NewObjectID(),
		IdentifierUsed: "user2",
		Type:           TypeTelegram,
		Code:           "654321",
		Success:        false,
		AttemptLeft:    2,
		TTL:            AttemptTTL,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	repoImpl.Create(ctx, attempt)
	id := attempt.ID.Hex()

	t.Run("gets auth attempt by ID", func(t *testing.T) {
		res, err := repoImpl.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, id, res.ID.Hex())
	})

	t.Run("returns error for invalid ID", func(t *testing.T) {
		res, err := repoImpl.GetByID(ctx, "bad-id")
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrInvalidAuthAttemptID)
	})
}

func TestAuthAttemptRepositoryImpl_Update(t *testing.T) {
	db, disconnect := testutils.SetupMongo(t)
	repoImpl := NewAuthAttemptRepository(db)
	defer disconnect()

	ctx := context.Background()
	attempt := AuthAttempt{
		ID:             primitive.NewObjectID(),
		IdentifierUsed: "user3",
		Type:           TypeTelegram,
		Code:           "999888",
		Success:        false,
		AttemptLeft:    3,
		TTL:            AttemptTTL,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	repoImpl.Create(ctx, attempt)
	id := attempt.ID.Hex()

	t.Run("updates attempt left", func(t *testing.T) {
		two := 2
		err := repoImpl.Update(ctx, id, UpdateAuthAttemptDTO{AttemptLeft: &two})
		assert.NoError(t, err)
	})

	t.Run("does not update on empty DTO", func(t *testing.T) {
		err := repoImpl.Update(ctx, id, UpdateAuthAttemptDTO{})
		assert.NoError(t, err)
	})

	t.Run("returns error on invalid ID", func(t *testing.T) {
		err := repoImpl.Update(ctx, "invalid-id", UpdateAuthAttemptDTO{})
		assert.ErrorIs(t, err, ErrInvalidAuthAttemptID)
	})
}

func TestAuthAttemptRepositoryImpl_Delete(t *testing.T) {
	db, disconnect := testutils.SetupMongo(t)
	repoImpl := NewAuthAttemptRepository(db)
	defer disconnect()

	ctx := context.Background()
	attempt := AuthAttempt{
		ID:             primitive.NewObjectID(),
		IdentifierUsed: "user4",
		Type:           TypeTelegram,
		Code:           "777666",
		Success:        false,
		AttemptLeft:    1,
		TTL:            AttemptTTL,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	repoImpl.Create(ctx, attempt)
	id := attempt.ID.Hex()

	t.Run("deletes auth attempt", func(t *testing.T) {
		err := repoImpl.Delete(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("delete returns error for bad ID", func(t *testing.T) {
		err := repoImpl.Delete(ctx, "bad-id")
		assert.ErrorIs(t, err, ErrInvalidAuthAttemptID)
	})

	t.Run("delete returns not found for missing attempt", func(t *testing.T) {
		fake := primitive.NewObjectID().Hex()
		err := repoImpl.Delete(ctx, fake)
		assert.ErrorIs(t, err, ErrAuthAttemptNotFound)
	})
}

func TestAuthAttemptRepositoryImpl_GetByTelegramUsername(t *testing.T) {
	db, disconnect := testutils.SetupMongo(t)
	repoImpl := NewAuthAttemptRepository(db)
	defer disconnect()

	ctx := context.Background()
	username := "teleuser"

	attempt := AuthAttempt{
		ID:             primitive.NewObjectID(),
		IdentifierUsed: username,
		Type:           TypeTelegram,
		Code:           "000111",
		Success:        false,
		AttemptLeft:    2,
		TTL:            AttemptTTL,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	_, err := repoImpl.Create(ctx, attempt)
	assert.NoError(t, err)

	t.Run("returns valid attempt", func(t *testing.T) {
		res, err := repoImpl.GetByTelegramUsername(ctx, username)
		assert.NoError(t, err)
		assert.Equal(t, username, res.IdentifierUsed)
	})

	t.Run("returns error if no valid attempt", func(t *testing.T) {
		expired := attempt
		expired.ID = primitive.NewObjectID()
		expired.CreatedAt = time.Now().Add(-10 * time.Minute)
		expired.AttemptLeft = 1
		_, err := repoImpl.Create(ctx, expired)
		assert.NoError(t, err)

		res, err := repoImpl.GetByTelegramUsername(ctx, "nonexistent")
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrAuthAttemptNotFound)
	})
}
