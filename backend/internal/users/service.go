package users

import (
	"context"
)

// Service defines user-related business logic.
//
// It provides a layer over the Repository abstraction, allowing validation,
// additional business rules, and future extensions (e.g., notifications, acls, airdrops).
type Service interface {
	// CreateUser adds a new user to the system.
	// Returns ErrUserExists if a user with the same Telegram ID already exists.
	CreateUser(ctx context.Context, dto CreateUserDTO) (*ReadUserDTO, error)

	// GetUserByID fetches a user by their MongoDB ObjectID string.
	// Returns ErrUserNotFound or ErrInvalidIDFormat if not found or invalid.
	GetUserByID(ctx context.Context, id string) (*ReadUserDTO, error)

	// GetUserByTelegramID finds a user using their Telegram ID.
	// Returns ErrUserNotFound if no user exists with the given Telegram ID.
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*ReadUserDTO, error)

	// UpdateUser modifies a user’s data by ID using the provided update DTO.
	// Returns ErrUserNotFound or ErrInvalidIDFormat if applicable.
	UpdateUser(ctx context.Context, id string, dto UpdateUserDTO) error

	// DeleteUser performs a soft delete (marks deleted_at) for a user by ID.
	// Returns ErrUserNotFound or ErrInvalidIDFormat if applicable.
	DeleteUser(ctx context.Context, id string) error
}

// service is the default implementation of the Service interface.
type service struct {
	repo Repository
}

// NewService creates a new user service using the provided Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateUser(ctx context.Context, dto CreateUserDTO) (*ReadUserDTO, error) {
	return s.repo.Create(ctx, dto)
}

func (s *service) GetUserByID(ctx context.Context, id string) (*ReadUserDTO, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetUserByTelegramID(ctx context.Context, telegramID int64) (*ReadUserDTO, error) {
	return s.repo.GetByTelegramID(ctx, telegramID)
}

func (s *service) UpdateUser(ctx context.Context, id string, dto UpdateUserDTO) error {
	return s.repo.UpdateByID(ctx, id, dto)
}

func (s *service) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteByID(ctx, id)
}
