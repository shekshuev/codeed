package users

import (
	"context"
	"errors"
)

var (
	// ErrUserNotFound is returned when a user with the given ID or Telegram ID is not found in the database.
	ErrUserNotFound = errors.New("user not found")

	// ErrUserExists is returned when attempting to create a user that already exists (based on Telegram ID).
	ErrUserExists = errors.New("user already exists")
)

// Repository defines a persistence interface for working with users.
// It abstracts away the storage layer (e.g., MongoDB).
type Repository interface {
	// Create adds a new user to the storage based on the provided CreateUserDTO.
	// Returns ErrUserExists if a user with the same Telegram ID already exists.
	Create(ctx context.Context, dto CreateUserDTO) (*ReadUserDTO, error)

	// GetByID retrieves a user by their MongoDB ObjectID string.
	// Returns ErrUserNotFound if the user does not exist or ID is invalid.
	GetByID(ctx context.Context, id string) (*ReadUserDTO, error)

	// GetByTelegramID finds a user by their Telegram user ID.
	// Returns ErrUserNotFound if no user with that Telegram ID exists.
	GetByTelegramID(ctx context.Context, telegramID int64) (*ReadUserDTO, error)

	// UpdateByID modifies a user's fields based on UpdateUserDTO.
	// Only non-nil fields in DTO are updated.
	// Returns ErrUserNotFound if no user with that ID exists.
	UpdateByID(ctx context.Context, id string, dto UpdateUserDTO) error
}
