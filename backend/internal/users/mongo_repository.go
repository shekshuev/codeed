package users

import (
	"context"
	"errors"
	"time"

	"github.com/shekshuev/codeed/backend/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoRepository provides MongoDB-based implementation of the user Repository interface.
type UserRepositoryImpl struct {
	collection *mongo.Collection
	log        *logger.Logger
}

// NewMongoRepository creates a new MongoRepository bound to the "users" collection.
func NewUserRepository(db *mongo.Database) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		collection: db.Collection("users"),
		log:        logger.NewLogger(),
	}
}

// Create inserts a new user document based on the provided CreateUserDTO.
// It first checks if a user with the same Telegram ID already exists.
// Returns ErrUserExists if duplicate, or the created user as ReadUserDTO.
func (r *UserRepositoryImpl) Create(ctx context.Context, dto CreateUserDTO) (*ReadUserDTO, error) {
	r.log.Sugar.Infof("Attempting to create user: telegram_id=%d", dto.TelegramID)
	count, err := r.collection.CountDocuments(ctx, bson.M{"telegram_id": dto.TelegramID})
	if err != nil {
		r.log.Sugar.Errorw("Failed to check for existing user", "error", err)
		return nil, err
	}
	if count > 0 {
		r.log.Sugar.Warnf("User already exists: telegram_id=%d", dto.TelegramID)
		return nil, ErrUserExists
	}

	user := dto.ToUserFromCreateDTO()
	_, err = r.collection.InsertOne(ctx, user)
	if err != nil {
		r.log.Sugar.Errorw("Failed to insert user", "error", err)
		return nil, err
	}
	readDto := user.ToReadUserDTO()
	r.log.Sugar.Infof("User created: id=%s", user.ID.Hex())
	return readDto, nil
}

// GetByID fetches a user by their string MongoDB ObjectID.
// Returns ErrUserNotFound if no user exists with the given ID,
// or an error if the ID format is invalid or DB error occurs.
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id string) (*ReadUserDTO, error) {
	r.log.Sugar.Infof("Fetching user by ID: %s", id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnw("Invalid user ID format", "id", id)
		return nil, ErrInvalidIDFormat
	}

	var u User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		r.log.Sugar.Warnf("User not found: id=%s", id)
		return nil, ErrUserNotFound
	} else if err != nil {
		r.log.Sugar.Errorw("Failed to fetch user", "error", err)
		return nil, err
	}

	return u.ToReadUserDTO(), nil
}

// GetByTelegramID finds a user by their Telegram ID (unique external identity).
// Returns ErrUserNotFound if no user exists with the given Telegram ID.
func (r *UserRepositoryImpl) GetByTelegramID(ctx context.Context, telegramID int64) (*ReadUserDTO, error) {
	r.log.Sugar.Infof("Fetching user by telegram_id: %d", telegramID)

	var u User
	err := r.collection.FindOne(ctx, bson.M{"telegram_id": telegramID}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		r.log.Sugar.Warnf("User not found by telegram_id: %d", telegramID)
		return nil, ErrUserNotFound
	} else if err != nil {
		r.log.Sugar.Errorw("Failed to fetch user by telegram_id", "error", err)
		return nil, err
	}

	return u.ToReadUserDTO(), nil
}

// UpdateByID updates a user by Mongo ObjectID string using non-nil fields from UpdateUserDTO.
// Returns ErrUserNotFound if no such user exists. Does nothing if DTO is empty.
func (r *UserRepositoryImpl) UpdateByID(ctx context.Context, id string, dto UpdateUserDTO) error {
	r.log.Sugar.Infof("Updating user: id=%s", id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnw("Invalid user ID format", "id", id)
		return ErrInvalidIDFormat
	}

	update := dto.ToBsonUpdateFromUpdateDTO()
	if len(update) == 0 {
		r.log.Sugar.Infof("No fields to update for user id=%s", id)
		return nil
	}

	res, err := r.collection.UpdateByID(ctx, objectID, update)
	if err != nil {
		r.log.Sugar.Errorw("Failed to update user", "id", id, "error", err)
		return err
	}
	if res.MatchedCount == 0 {
		r.log.Sugar.Warnf("User not found for update: id=%s", id)
		return ErrUserNotFound
	}

	r.log.Sugar.Infof("User updated successfully: id=%s", id)
	return nil
}

// DeleteByID marks a user as deleted by their MongoDB ObjectID string.
// Returns ErrUserNotFound if no user exists with the given ID.
func (r *UserRepositoryImpl) DeleteByID(ctx context.Context, id string) error {
	r.log.Sugar.Infof("Soft deleting user: id=%s", id)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnw("Invalid user ID format", "id", id)
		return ErrInvalidIDFormat
	}

	res, err := r.collection.UpdateByID(ctx, objectID, bson.M{
		"$set": bson.M{"deleted_at": time.Now().UTC()},
	})
	if err != nil {
		r.log.Sugar.Errorw("Failed to soft delete user", "id", id, "error", err)
		return err
	}
	if res.ModifiedCount == 0 {
		r.log.Sugar.Warnf("User not found for soft delete: id=%s", id)
		return ErrUserNotFound
	}

	r.log.Sugar.Infof("User soft deleted successfully: id=%s", id)
	return nil
}
