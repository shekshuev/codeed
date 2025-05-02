package auth

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Predefined errors for auth attempt operations.
var (
	ErrAuthAttemptNotFound  = errors.New("auth attempt not found")
	ErrInvalidAuthAttemptID = errors.New("invalid auth attempt ID")
)

// AuthAttemptRepositoryImpl provides MongoDB-based implementation
// for storing and retrieving authentication attempts.
type AuthAttemptRepositoryImpl struct {
	collection *mongo.Collection
}

// NewAuthAttemptRepository creates a new instance of the repository
// using the provided MongoDB database connection.
func NewAuthAttemptRepository(db *mongo.Database) *AuthAttemptRepositoryImpl {
	return &AuthAttemptRepositoryImpl{
		collection: db.Collection("auth_attempts"),
	}
}

// Create inserts a new auth attempt document into the collection.
func (r *AuthAttemptRepositoryImpl) Create(ctx context.Context, attempt AuthAttempt) (*AuthAttempt, error) {
	_, err := r.collection.InsertOne(ctx, attempt)
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

// GetByID fetches a single auth attempt by its ID.
// Returns ErrInvalidAuthAttemptID or ErrAuthAttemptNotFound if not found.
func (r *AuthAttemptRepositoryImpl) GetByID(ctx context.Context, id string) (*AuthAttempt, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidAuthAttemptID
	}

	var attempt AuthAttempt
	err = r.collection.FindOne(ctx, bson.M{
		"_id": objID,
	}).Decode(&attempt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrAuthAttemptNotFound
		}
		return nil, err
	}
	return &attempt, nil
}

// Update modifies specific fields (success, attempt_left) of the given auth attempt.
// Returns ErrAuthAttemptNotFound if the document doesn't exist.
// Returns silently if the DTO has no fields to update.
func (r *AuthAttemptRepositoryImpl) Update(ctx context.Context, id string, dto UpdateAuthAttemptDTO) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidAuthAttemptID
	}

	update := bson.M{}
	if dto.Success != nil {
		update["success"] = *dto.Success
	}
	if dto.AttemptLeft != nil {
		update["attempt_left"] = *dto.AttemptLeft
	}
	if len(update) == 0 {
		return nil // nothing to update
	}
	update["updated_at"] = time.Now().UTC()

	res, err := r.collection.UpdateOne(ctx, bson.M{
		"_id": objID,
	}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrAuthAttemptNotFound
	}
	return nil
}

// Delete removes the auth attempt from the collection permanently.
// Returns ErrInvalidAuthAttemptID or ErrAuthAttemptNotFound as needed.
func (r *AuthAttemptRepositoryImpl) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidAuthAttemptID
	}

	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrAuthAttemptNotFound
	}
	return nil
}
