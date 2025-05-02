package auth

import (
	"context"
	"testing"
	"time"

	"github.com/shekshuev/codeed/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAuthAttemptRepository_CRUD(t *testing.T) {
	db, disconnect := testutils.SetupMongo(t)
	repo := NewAuthAttemptRepository(db)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var attemptID string

	t.Run("creates an auth attempt", func(t *testing.T) {
		attempt := AuthAttempt{
			ID:             primitive.NewObjectID(),
			IdentifierUsed: "telegram_user",
			Type:           TypeTelegram,
			Code:           "123456",
			Success:        false,
			AttemptLeft:    3,
			TTL:            5 * time.Minute,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}

		created, err := repo.Create(ctx, attempt)
		assert.NoError(t, err)
		assert.Equal(t, attempt.ID, created.ID)
		attemptID = created.ID.Hex()
	})

	t.Run("retrieves an auth attempt by ID", func(t *testing.T) {
		found, err := repo.GetByID(ctx, attemptID)
		assert.NoError(t, err)
		assert.Equal(t, "telegram_user", found.IdentifierUsed)
	})

	t.Run("updates attempt_left", func(t *testing.T) {
		newLeft := 2
		err := repo.Update(ctx, attemptID, UpdateAuthAttemptDTO{
			AttemptLeft: &newLeft,
		})
		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, attemptID)
		assert.NoError(t, err)
		assert.Equal(t, 2, updated.AttemptLeft)
	})

	t.Run("deletes an auth attempt", func(t *testing.T) {
		err := repo.Delete(ctx, attemptID)
		assert.NoError(t, err)

		_, err = repo.GetByID(ctx, attemptID)
		assert.ErrorIs(t, err, ErrAuthAttemptNotFound)
	})
}

func TestAuthAttemptRepository_InvalidID(t *testing.T) {
	db, disconnect := testutils.SetupMongo(t)
	repo := NewAuthAttemptRepository(db)
	defer disconnect()

	ctx := context.Background()

	t.Run("returns error for invalid GetByID", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "invalid-id")
		assert.ErrorIs(t, err, ErrInvalidAuthAttemptID)
	})

	t.Run("returns error for invalid Update", func(t *testing.T) {
		err := repo.Update(ctx, "invalid-id", UpdateAuthAttemptDTO{})
		assert.ErrorIs(t, err, ErrInvalidAuthAttemptID)
	})

	t.Run("returns error for invalid Delete", func(t *testing.T) {
		err := repo.Delete(ctx, "invalid-id")
		assert.ErrorIs(t, err, ErrInvalidAuthAttemptID)
	})
}
