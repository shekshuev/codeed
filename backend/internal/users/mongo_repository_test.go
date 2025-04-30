package users

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupMongo(t *testing.T) (*MongoRepository, func()) {
	ctx := context.Background()

	container, err := mongodb.RunContainer(ctx,
		testcontainers.WithImage("mongo:6"),
	)
	assert.NoError(t, err)

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	uri, err := container.ConnectionString(ctx)
	assert.NoError(t, err)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	assert.NoError(t, err)

	db := client.Database("testdb")

	return NewMongoRepository(db), func() {
		_ = client.Disconnect(ctx)
	}
}

func TestMongoRepository_Create(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("successfully creates user", func(t *testing.T) {
		dto := CreateUserDTO{
			TelegramID: 1111,
			Username:   "testuser",
			FirstName:  "Test",
			LastName:   "User",
			Role:       "student",
		}

		user, err := repo.Create(ctx, dto)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, dto.TelegramID, user.TelegramID)
		assert.Equal(t, dto.Username, user.Username)
		assert.Equal(t, dto.FirstName, user.FirstName)
		assert.Equal(t, dto.LastName, user.LastName)
		assert.Equal(t, dto.Role, user.Role)
		assert.NotEmpty(t, user.ID)
		assert.NotEmpty(t, user.CreatedAt)
	})

	t.Run("returns error on duplicate TelegramID", func(t *testing.T) {
		dto := CreateUserDTO{
			TelegramID: 1111,
			Username:   "otheruser",
			FirstName:  "Dup",
			LastName:   "User",
			Role:       "student",
		}

		user, err := repo.Create(ctx, dto)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, ErrUserExists)
	})
}

func TestMongoRepository_GetByID(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("returns user by ID", func(t *testing.T) {
		dto := CreateUserDTO{
			TelegramID: 2222,
			Username:   "getbyid",
			FirstName:  "Get",
			LastName:   "Test",
			Role:       "admin",
		}

		created, err := repo.Create(ctx, dto)
		assert.NoError(t, err)

		found, err := repo.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Username, found.Username)
	})

	t.Run("returns error if ID format is invalid", func(t *testing.T) {
		user, err := repo.GetByID(ctx, "invalid-id")
		assert.Nil(t, user)
		assert.ErrorIs(t, err, ErrInvalidIDFormat)
	})

	t.Run("returns error if user not found", func(t *testing.T) {
		fakeID := primitive.NewObjectID().Hex()
		user, err := repo.GetByID(ctx, fakeID)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestMongoRepository_GetByTelegramID(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("returns user by Telegram ID", func(t *testing.T) {
		dto := CreateUserDTO{
			TelegramID: 3333,
			Username:   "bytelegram",
			FirstName:  "Tele",
			LastName:   "Gram",
			Role:       "student",
		}

		created, err := repo.Create(ctx, dto)
		assert.NoError(t, err)

		found, err := repo.GetByTelegramID(ctx, dto.TelegramID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Username, found.Username)
	})

	t.Run("returns error if user with Telegram ID not found", func(t *testing.T) {
		user, err := repo.GetByTelegramID(ctx, 999999999)
		assert.Nil(t, user)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestMongoRepository_UpdateByID(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("updates existing user fields", func(t *testing.T) {
		created, err := repo.Create(ctx, CreateUserDTO{
			TelegramID: 4444,
			Username:   "beforeupdate",
			FirstName:  "Old",
			LastName:   "Name",
			Role:       "student",
		})
		assert.NoError(t, err)

		newUsername := "afterupdate"
		newFirstName := "Newname"
		dto := UpdateUserDTO{
			Username:  &newUsername,
			FirstName: &newFirstName,
		}

		err = repo.UpdateByID(ctx, created.ID, dto)
		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, newUsername, updated.Username)
		assert.Equal(t, newFirstName, updated.FirstName)
		assert.Equal(t, created.LastName, updated.LastName)
	})

	t.Run("does nothing if DTO is empty", func(t *testing.T) {
		created, err := repo.Create(ctx, CreateUserDTO{
			TelegramID: 5555,
			Username:   "unchanged",
			FirstName:  "Still",
			LastName:   "Here",
			Role:       "student",
		})
		assert.NoError(t, err)

		err = repo.UpdateByID(ctx, created.ID, UpdateUserDTO{})
		assert.NoError(t, err)

		same, err := repo.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.Username, same.Username)
	})

	t.Run("returns error if ID is invalid", func(t *testing.T) {
		newUsername := "fail"
		dto := UpdateUserDTO{Username: &newUsername}

		err := repo.UpdateByID(ctx, "invalid_id", dto)
		assert.ErrorIs(t, err, ErrInvalidIDFormat)
	})

	t.Run("returns error if user not found", func(t *testing.T) {
		newUsername := "notfound"
		dto := UpdateUserDTO{Username: &newUsername}

		fakeID := primitive.NewObjectID().Hex()
		err := repo.UpdateByID(ctx, fakeID, dto)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestMongoRepository_DeleteByID(t *testing.T) {
	repo, disconnect := setupMongo(t)
	defer disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("successfully soft deletes user", func(t *testing.T) {
		created, err := repo.Create(ctx, CreateUserDTO{
			TelegramID: 6666,
			Username:   "tobedeleted",
			FirstName:  "Gone",
			LastName:   "Soon",
			Role:       "student",
		})
		assert.NoError(t, err)

		err = repo.DeleteByID(ctx, created.ID)
		assert.NoError(t, err)

		objID, _ := primitive.ObjectIDFromHex(created.ID)
		var raw bson.M
		err = repo.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&raw)
		assert.NoError(t, err)

		deletedAt, ok := raw["deleted_at"].(primitive.DateTime)
		assert.True(t, ok)
		goTime := deletedAt.Time()
		assert.WithinDuration(t, time.Now().UTC(), goTime, 5*time.Second)

	})

	t.Run("returns error if ID is invalid", func(t *testing.T) {
		err := repo.DeleteByID(ctx, "invalid_id")
		assert.ErrorIs(t, err, ErrInvalidIDFormat)
	})

	t.Run("returns error if user not found", func(t *testing.T) {
		fakeID := primitive.NewObjectID().Hex()
		err := repo.DeleteByID(ctx, fakeID)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}
