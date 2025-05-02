package auth

import (
	"context"
	"errors"
	"time"

	"github.com/shekshuev/codeed/backend/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Predefined errors for auth attempt operations.
var (
	ErrAuthAttemptNotFound  = errors.New("auth attempt not found")
	ErrInvalidAuthAttemptID = errors.New("invalid auth attempt ID")
)

type AuthAttemptRepositoryImpl struct {
	collection *mongo.Collection
	log        *logger.Logger
}

// NewAuthAttemptRepository creates a new instance of the repository
// using the provided MongoDB database connection.
func NewAuthAttemptRepository(db *mongo.Database) *AuthAttemptRepositoryImpl {
	return &AuthAttemptRepositoryImpl{
		collection: db.Collection("auth_attempts"),
		log:        logger.NewLogger(),
	}
}

// Create inserts a new auth attempt document into the collection.
func (r *AuthAttemptRepositoryImpl) Create(ctx context.Context, attempt AuthAttempt) (*AuthAttempt, error) {
	_, err := r.collection.InsertOne(ctx, attempt)
	if err != nil {
		r.log.Sugar.Errorf("Failed to create auth attempt for '%s': %v", attempt.IdentifierUsed, err)
		return nil, err
	}
	r.log.Sugar.Infof("Created auth attempt: id=%s identifier=%s", attempt.ID.Hex(), attempt.IdentifierUsed)
	return &attempt, nil
}

// GetByID fetches a single auth attempt by its ID.
func (r *AuthAttemptRepositoryImpl) GetByID(ctx context.Context, id string) (*AuthAttempt, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnf("Invalid auth attempt ID format: %s", id)
		return nil, ErrInvalidAuthAttemptID
	}

	var attempt AuthAttempt
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&attempt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.log.Sugar.Infof("Auth attempt not found: id=%s", id)
			return nil, ErrAuthAttemptNotFound
		}
		r.log.Sugar.Errorf("Error retrieving auth attempt %s: %v", id, err)
		return nil, err
	}

	r.log.Sugar.Debugf("Retrieved auth attempt: id=%s", id)
	return &attempt, nil
}

// Update modifies specific fields (success, attempt_left) of the given auth attempt.
func (r *AuthAttemptRepositoryImpl) Update(ctx context.Context, id string, dto UpdateAuthAttemptDTO) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnf("Invalid auth attempt ID for update: %s", id)
		return ErrInvalidAuthAttemptID
	}

	update := bson.M{}
	if dto.Success {
		update["success"] = dto.Success
	}
	if dto.AttemptLeft != nil {
		update["attempt_left"] = *dto.AttemptLeft
	}
	if len(update) == 0 {
		r.log.Sugar.Infof("No fields to update for auth attempt id=%s", id)
		return nil
	}
	update["updated_at"] = time.Now().UTC()

	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		r.log.Sugar.Errorf("Failed to update auth attempt id=%s: %v", id, err)
		return err
	}
	if res.MatchedCount == 0 {
		r.log.Sugar.Infof("Auth attempt not found for update: id=%s", id)
		return ErrAuthAttemptNotFound
	}

	r.log.Sugar.Infof("Updated auth attempt: id=%s fields=%v", id, update)
	return nil
}

// Delete removes the auth attempt from the collection permanently.
func (r *AuthAttemptRepositoryImpl) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		r.log.Sugar.Warnf("Invalid auth attempt ID for delete: %s", id)
		return ErrInvalidAuthAttemptID
	}

	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		r.log.Sugar.Errorf("Failed to delete auth attempt id=%s: %v", id, err)
		return err
	}
	if res.DeletedCount == 0 {
		r.log.Sugar.Infof("Auth attempt not found for deletion: id=%s", id)
		return ErrAuthAttemptNotFound
	}

	r.log.Sugar.Infof("Deleted auth attempt: id=%s", id)
	return nil
}

// GetByTelegramUsername retrieves the most recent auth attempt for the given Telegram username,
// only if the attempt is still active (has remaining attempts and TTL not expired).
//
// It returns ErrAuthAttemptNotFound if no valid attempt exists.
func (r *AuthAttemptRepositoryImpl) GetByTelegramUsername(ctx context.Context, telegramUsername string) (*AuthAttempt, error) {
	now := time.Now().UTC()

	filter := bson.M{
		"identifier_used": telegramUsername,
		"type":            TypeTelegram,
		"attempt_left":    bson.M{"$gt": 0},
		"created_at":      bson.M{"$gte": now.Add(-AttemptTTL)},
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var attempt AuthAttempt
	err := r.collection.FindOne(ctx, filter, opts).Decode(&attempt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.log.Sugar.Infof("No valid auth attempt found for telegram user: %s", telegramUsername)
			return nil, ErrAuthAttemptNotFound
		}
		r.log.Sugar.Errorf("Failed to fetch auth attempt by telegram user %s: %v", telegramUsername, err)
		return nil, err
	}

	r.log.Sugar.Infof("Found valid auth attempt for telegram user: %s (id=%s)", telegramUsername, attempt.ID.Hex())
	return &attempt, nil
}
