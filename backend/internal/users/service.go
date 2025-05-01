package users

import (
	"context"

	"github.com/shekshuev/codeed/backend/internal/logger"
)

// Service defines user-related business logic.
//
// It provides a layer over the Repository abstraction, allowing validation,
// additional business rules, and future extensions (e.g., notifications, acls, airdrops).
type UserService interface {
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
type UserServiceImpl struct {
	repo UserRepository
	log  *logger.Logger
}

// NewService creates a new user service using the provided Repository.
func NewService(repo UserRepository) UserService {
	return &UserServiceImpl{repo: repo, log: logger.NewLogger()}
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, dto CreateUserDTO) (*ReadUserDTO, error) {
	s.log.Sugar.Infof("Creating user: telegram_id=%d username=%s", dto.TelegramID, dto.Username)
	user, err := s.repo.Create(ctx, dto)
	if err != nil {
		s.log.Sugar.Warnw("Failed to create user", "telegram_id", dto.TelegramID, "error", err)
		return nil, err
	}
	s.log.Sugar.Infof("User created: id=%s", user.ID)
	return user, nil
}

func (s *UserServiceImpl) GetUserByID(ctx context.Context, id string) (*ReadUserDTO, error) {
	s.log.Sugar.Infof("Fetching user by ID: %s", id)
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Sugar.Warnw("Failed to fetch user by ID", "id", id, "error", err)
		return nil, err
	}
	return user, nil
}

func (s *UserServiceImpl) GetUserByTelegramID(ctx context.Context, telegramID int64) (*ReadUserDTO, error) {
	s.log.Sugar.Infof("Fetching user by Telegram ID: %d", telegramID)
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		s.log.Sugar.Warnw("Failed to fetch user by Telegram ID", "telegram_id", telegramID, "error", err)
		return nil, err
	}
	return user, nil
}

func (s *UserServiceImpl) UpdateUser(ctx context.Context, id string, dto UpdateUserDTO) error {
	s.log.Sugar.Infof("Updating user: id=%s", id)
	err := s.repo.UpdateByID(ctx, id, dto)
	if err != nil {
		s.log.Sugar.Warnw("Failed to update user", "id", id, "error", err)
		return err
	}
	s.log.Sugar.Infof("User updated: id=%s", id)
	return nil
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, id string) error {
	s.log.Sugar.Infof("Deleting user: id=%s", id)
	err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		s.log.Sugar.Warnw("Failed to delete user", "id", id, "error", err)
		return err
	}
	s.log.Sugar.Infof("User deleted: id=%s", id)
	return nil
}
