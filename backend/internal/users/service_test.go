package users

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_CreateUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	service := NewService(mockRepo)

	createDTO := CreateUserDTO{
		TelegramID: 123,
		Username:   "john",
		FirstName:  "John",
		LastName:   "Doe",
		Role:       "student",
	}
	readDTO := &ReadUserDTO{
		ID:         "abc123",
		TelegramID: 123,
		Username:   "john",
		FirstName:  "John",
		LastName:   "Doe",
		Role:       "student",
		CreatedAt:  "2024-01-01T00:00:00Z",
	}

	t.Run("successful creation", func(t *testing.T) {
		mockRepo.
			EXPECT().
			Create(ctx, createDTO).
			Return(readDTO, nil).
			Times(1)

		result, err := service.CreateUser(ctx, createDTO)
		assert.NoError(t, err)
		assert.Equal(t, readDTO, result)
	})

	t.Run("duplicate user", func(t *testing.T) {
		mockRepo.
			EXPECT().
			Create(ctx, createDTO).
			Return(nil, ErrUserExists).
			Times(1)

		result, err := service.CreateUser(ctx, createDTO)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrUserExists)
	})
}

func TestService_GetUserByID(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	service := NewService(mockRepo)

	userID := "abc123"
	readDTO := &ReadUserDTO{
		ID:         userID,
		TelegramID: 123,
		Username:   "john",
		FirstName:  "John",
		LastName:   "Doe",
		Role:       "student",
		CreatedAt:  "2024-01-01T00:00:00Z",
	}

	t.Run("returns user successfully", func(t *testing.T) {
		mockRepo.
			EXPECT().
			GetByID(ctx, userID).
			Return(readDTO, nil).
			Times(1)

		result, err := service.GetUserByID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, readDTO, result)
	})

	t.Run("returns error if user not found", func(t *testing.T) {
		mockRepo.
			EXPECT().
			GetByID(ctx, userID).
			Return(nil, ErrUserNotFound).
			Times(1)

		result, err := service.GetUserByID(ctx, userID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestService_GetUserByTelegramID(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	service := NewService(mockRepo)

	telegramID := int64(987654321)
	readDTO := &ReadUserDTO{
		ID:         "def456",
		TelegramID: telegramID,
		Username:   "telegram_user",
		FirstName:  "Tele",
		LastName:   "User",
		Role:       "student",
		CreatedAt:  "2024-01-02T00:00:00Z",
	}

	t.Run("returns user by Telegram ID", func(t *testing.T) {
		mockRepo.
			EXPECT().
			GetByTelegramID(ctx, telegramID).
			Return(readDTO, nil).
			Times(1)

		result, err := service.GetUserByTelegramID(ctx, telegramID)
		assert.NoError(t, err)
		assert.Equal(t, readDTO, result)
	})

	t.Run("returns error if Telegram user not found", func(t *testing.T) {
		mockRepo.
			EXPECT().
			GetByTelegramID(ctx, telegramID).
			Return(nil, ErrUserNotFound).
			Times(1)

		result, err := service.GetUserByTelegramID(ctx, telegramID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	service := NewService(mockRepo)

	userID := "abc123"
	username := "updateduser"
	updateDTO := UpdateUserDTO{
		Username: &username,
	}

	t.Run("successfully updates user", func(t *testing.T) {
		mockRepo.
			EXPECT().
			UpdateByID(ctx, userID, updateDTO).
			Return(nil).
			Times(1)

		err := service.UpdateUser(ctx, userID, updateDTO)
		assert.NoError(t, err)
	})

	t.Run("returns error if user not found", func(t *testing.T) {
		mockRepo.
			EXPECT().
			UpdateByID(ctx, userID, updateDTO).
			Return(ErrUserNotFound).
			Times(1)

		err := service.UpdateUser(ctx, userID, updateDTO)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestService_DeleteUser(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	service := NewService(mockRepo)

	userID := "abc123"

	t.Run("successfully soft deletes user", func(t *testing.T) {
		mockRepo.
			EXPECT().
			DeleteByID(ctx, userID).
			Return(nil).
			Times(1)

		err := service.DeleteUser(ctx, userID)
		assert.NoError(t, err)
	})

	t.Run("returns error if user not found", func(t *testing.T) {
		mockRepo.
			EXPECT().
			DeleteByID(ctx, userID).
			Return(ErrUserNotFound).
			Times(1)

		err := service.DeleteUser(ctx, userID)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}
