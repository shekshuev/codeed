package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shekshuev/codeed/backend/internal/config"
	"github.com/shekshuev/codeed/backend/internal/users"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestService_RequestTelegramCode(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockAuthAttemptRepository(ctrl)
	userService := users.NewMockUserService(ctrl)
	cfg := &config.Config{}
	service := NewAuthService(repo, userService, cfg)

	dto := CreateTelegramCodeRequestDTO{TelegramUsername: "user1"}
	attempt := dto.ToAuthAttemptFromCreateTelegramCodeRequestDTO()

	t.Run("creates new auth attempt for new user", func(t *testing.T) {
		repo.EXPECT().GetByTelegramUsername(ctx, "user1").Return(nil, errors.New("not found"))
		repo.EXPECT().Create(ctx, gomock.Any()).Return(&attempt, nil)

		res, err := service.RequestTelegramCode(ctx, dto)
		assert.NoError(t, err)
		assert.Equal(t, "user1", res.TelegramUsername)
	})

	t.Run("returns error if attempt already exists", func(t *testing.T) {
		repo.EXPECT().GetByTelegramUsername(ctx, "user1").Return(&attempt, nil)
		res, err := service.RequestTelegramCode(ctx, dto)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrAuthAttemptAlreadyExists)
	})

	t.Run("returns error if create fails", func(t *testing.T) {
		repo.EXPECT().GetByTelegramUsername(ctx, "user2").Return(nil, errors.New("not found"))
		repo.EXPECT().Create(ctx, gomock.Any()).Return(nil, errors.New("insert failed"))

		dto2 := CreateTelegramCodeRequestDTO{TelegramUsername: "user2"}
		res, err := service.RequestTelegramCode(ctx, dto2)
		assert.Nil(t, res)
		assert.Error(t, err)
	})
}

func TestService_CheckTelegramCode(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockAuthAttemptRepository(ctrl)
	userService := users.NewMockUserService(ctrl)
	cfg := &config.Config{
		AccessTokenSecret:   "secret",
		RefreshTokenSecret:  "secret",
		AccessTokenExpires:  time.Minute,
		RefreshTokenExpires: time.Hour,
	}
	service := NewAuthService(repo, userService, cfg)

	attempt := &AuthAttempt{
		ID:             primitiveObjectID("64e21f6a2f1c8e1c0a3d4f5b"),
		IdentifierUsed: "user1",
		Code:           "123456",
		AttemptLeft:    2,
		CreatedAt:      time.Now(),
	}
	readUser := &users.ReadUserDTO{ID: "u1"}

	t.Run("successfully checks code and returns tokens", func(t *testing.T) {
		repo.EXPECT().GetByID(ctx, "id1").Return(attempt, nil)
		repo.EXPECT().Update(ctx, attempt.ID.Hex(), gomock.Any()).Return(nil)
		userService.EXPECT().GetUserByTelegramUsername(ctx, "user1").Return(readUser, nil)

		dto := CreateCheckTelegramCodeRequestDTO{ID: "id1", TelegramUsername: "user1", Code: "123456"}
		pair, err := service.CheckTelegramCode(ctx, dto)
		assert.NoError(t, err)
		assert.NotEmpty(t, pair.AccessToken)
		assert.NotEmpty(t, pair.RefreshToken)
	})

	t.Run("username mismatch", func(t *testing.T) {
		wrong := *attempt
		wrong.IdentifierUsed = "another"
		repo.EXPECT().GetByID(ctx, "id2").Return(&wrong, nil)
		dto := CreateCheckTelegramCodeRequestDTO{ID: "id2", TelegramUsername: "user1", Code: "123456"}
		res, err := service.CheckTelegramCode(ctx, dto)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrWrongTelegramUsername)
	})

	t.Run("wrong code with no attempts left", func(t *testing.T) {
		noAttempt := *attempt
		noAttempt.AttemptLeft = 1
		noAttempt.Code = "000000"
		repo.EXPECT().GetByID(ctx, "id3").Return(&noAttempt, nil)
		repo.EXPECT().Delete(ctx, noAttempt.ID.Hex()).Return(nil)

		dto := CreateCheckTelegramCodeRequestDTO{ID: "id3", TelegramUsername: "user1", Code: "wrong"}
		res, err := service.CheckTelegramCode(ctx, dto)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, ErrInvalidCode)
	})

	t.Run("update fails on correct code", func(t *testing.T) {
		repo.EXPECT().GetByID(ctx, "id4").Return(attempt, nil)
		repo.EXPECT().Update(ctx, attempt.ID.Hex(), gomock.Any()).Return(errors.New("update error"))
		dto := CreateCheckTelegramCodeRequestDTO{ID: "id4", TelegramUsername: "user1", Code: "123456"}
		res, err := service.CheckTelegramCode(ctx, dto)
		assert.Nil(t, res)
		assert.Error(t, err)
	})
}

func primitiveObjectID(hex string) primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex(hex)
	return id
}
