package users

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoRepository provides MongoDB-based implementation of the user Repository interface.
type MongoRepository struct {
	collection *mongo.Collection
}

// NewMongoRepository creates a new MongoRepository bound to the "users" collection.
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection("users"),
	}
}

// Create inserts a new user document based on the provided CreateUserDTO.
// It first checks if a user with the same Telegram ID already exists.
// Returns ErrUserExists if duplicate, or the created user as ReadUserDTO.
func (r *MongoRepository) Create(ctx context.Context, dto CreateUserDTO) (*ReadUserDTO, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"telegram_id": dto.TelegramID})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrUserExists
	}

	user := dto.ToUserFromCreateDTO()
	_, err = r.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	readDto := user.ToReadUserDTO()
	return readDto, nil
}

// GetByID fetches a user by their string MongoDB ObjectID.
// Returns ErrUserNotFound if no user exists with the given ID,
// or an error if the ID format is invalid or DB error occurs.
func (r *MongoRepository) GetByID(ctx context.Context, id string) (*ReadUserDTO, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user id format")
	}
	var u User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	dto := u.ToReadUserDTO()
	return dto, nil
}

// GetByTelegramID finds a user by their Telegram ID (unique external identity).
// Returns ErrUserNotFound if no user exists with the given Telegram ID.
func (r *MongoRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*ReadUserDTO, error) {
	var u User
	err := r.collection.FindOne(ctx, bson.M{"telegram_id": telegramID}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	dto := u.ToReadUserDTO()
	return dto, nil
}

// UpdateByID updates a user by Mongo ObjectID string using non-nil fields from UpdateUserDTO.
// Returns ErrUserNotFound if no such user exists. Does nothing if DTO is empty.
func (r *MongoRepository) UpdateByID(ctx context.Context, id string, dto UpdateUserDTO) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user id format")
	}
	update := dto.ToBsonUpdateFromUpdateDTO()
	if len(update) == 0 {
		return nil
	}
	res, err := r.collection.UpdateByID(ctx, objectID, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}
