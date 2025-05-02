package auth

import (
	"context"
	"errors"

	"github.com/shekshuev/codeed/backend/internal/config"
	"github.com/shekshuev/codeed/backend/internal/logger"
	"github.com/shekshuev/codeed/backend/internal/users"
	"github.com/shekshuev/codeed/backend/internal/utils"
)

// ErrAuthAttemptAlreadyExists is returned when a code request is made for a username
// that already has a valid auth attempt in progress.
var ErrAuthAttemptAlreadyExists = errors.New("auth attempt already exists")

// ErrWrongTelegramUsername is returned when the provided Telegram username
// does not match the username associated with the auth attempt.
var ErrWrongTelegramUsername = errors.New("wrong telegram username")

// ErrInvalidCode is returned when the submitted verification code is incorrect
// or when the maximum number of allowed attempts has been exceeded.
var ErrInvalidCode = errors.New("invalid code")

// AuthService defines the interface for handling authentication
// attempts via Telegram code verification.
type AuthService interface {
	// RequestTelegramCode initiates a new auth attempt for the given Telegram username.
	// Returns a DTO with attempt ID and wait time, or an error if an active attempt already exists.
	RequestTelegramCode(ctx context.Context, dto CreateTelegramCodeRequestDTO) (*ReadTelegramCodeRequestDTO, error)

	// CheckTelegramCode validates the provided code against an existing auth attempt.
	// On success, returns a new pair of access and refresh tokens.
	CheckTelegramCode(ctx context.Context, dto CreateCheckTelegramCodeRequestDTO) (*TokenPairDTO, error)
}

// AuthServiceImpl implements AuthService using a repository,
// user service, config, and structured logging.
type AuthServiceImpl struct {
	repo        AuthAttemptRepository // Storage layer for auth attempts
	userService users.UserService     // External user service for user lookup
	log         *logger.Logger        // Structured logger
	cfg         *config.Config        // Application configuration
}

// NewAuthService constructs a new AuthServiceImpl with its dependencies injected.
func NewAuthService(
	repo AuthAttemptRepository,
	userService users.UserService,
	cfg *config.Config,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		repo:        repo,
		userService: userService,
		log:         logger.NewLogger(),
		cfg:         cfg,
	}
}

// RequestTelegramCode starts a new authentication attempt for a Telegram user.
// If a valid attempt already exists, it returns an ErrAuthAttemptAlreadyExists error.
func (s *AuthServiceImpl) RequestTelegramCode(
	ctx context.Context,
	dto CreateTelegramCodeRequestDTO,
) (*ReadTelegramCodeRequestDTO, error) {
	_, err := s.repo.GetByTelegramUsername(ctx, dto.TelegramUsername)
	if err == nil {
		s.log.Sugar.Infof("Telegram code request for existing user: %s", dto.TelegramUsername)
		return nil, ErrAuthAttemptAlreadyExists
	}

	attempt, err := s.repo.Create(ctx, dto.ToAuthAttemptFromCreateTelegramCodeRequestDTO())
	if err != nil {
		s.log.Sugar.Errorf("Failed to create auth attempt for '%s': %v", dto.TelegramUsername, err)
		return nil, err
	}

	return attempt.ToReadTelegramCodeRequestDTO(), nil
}

// CheckTelegramCode validates the verification code and issues JWT tokens if valid.
// It ensures the Telegram username matches, attempts are not exhausted,
// and the code is correct. On success, returns a new token pair.
func (s *AuthServiceImpl) CheckTelegramCode(
	ctx context.Context,
	dto CreateCheckTelegramCodeRequestDTO,
) (*TokenPairDTO, error) {
	attempt, err := s.repo.GetByID(ctx, dto.ID)
	if err != nil {
		s.log.Sugar.Errorf("Failed to get auth attempt for '%s': %v", dto.TelegramUsername, err)
		return nil, err
	}

	if attempt.IdentifierUsed != dto.TelegramUsername {
		s.log.Sugar.Warnf("Invalid identifier for auth attempt: %s", dto.TelegramUsername)
		return nil, ErrWrongTelegramUsername
	}

	err = s.performCodeCheck(ctx, attempt, dto.Code)
	if err != nil {
		s.log.Sugar.Errorf("Failed to check code for auth attempt: %s", dto.TelegramUsername)
		return nil, err
	}

	user, err := s.userService.GetUserByTelegramUsername(ctx, dto.TelegramUsername)
	if err != nil {
		s.log.Sugar.Errorf("Failed to get user by telegram username: %s", dto.TelegramUsername)
		return nil, err
	}

	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		s.log.Sugar.Errorf("Failed to generate token pair for user: %s", dto.TelegramUsername)
		return nil, err
	}

	s.log.Sugar.Infof("Successfully checked code for auth attempt: %s", dto.TelegramUsername)
	return tokenPair, nil
}

// performCodeCheck validates the code in the auth attempt.
// On success, it marks the attempt as successful. On failure,
// it either decrements the remaining attempts or deletes the attempt if exhausted.
func (s *AuthServiceImpl) performCodeCheck(
	ctx context.Context,
	attempt *AuthAttempt,
	code string,
) error {
	if attempt.AttemptLeft <= 0 {
		s.log.Sugar.Warnf("No attempts left for auth attempt: %s", attempt.ID)
		return ErrInvalidCode
	}

	if attempt.Code == code {
		err := s.repo.Update(ctx, attempt.ID.Hex(), UpdateAuthAttemptDTO{Success: true})
		if err != nil {
			s.log.Sugar.Errorf("Failed to update auth attempt for '%s': %v", attempt.ID, err)
			return err
		}
		return nil
	}

	s.log.Sugar.Warnf("Invalid code for auth attempt: %s", attempt.ID)
	attemptsLeft := attempt.AttemptLeft - 1

	if attemptsLeft == 0 {
		s.log.Sugar.Warnf("No attempts left for auth attempt: %s", attempt.ID)
		err := s.repo.Delete(ctx, attempt.ID.Hex())
		if err != nil {
			s.log.Sugar.Errorf("Failed to delete auth attempt for '%s': %v", attempt.ID, err)
			return err
		}
	} else {
		s.log.Sugar.Infof("Decrementing attempt: %d → %d", attempt.AttemptLeft, attemptsLeft)
		err := s.repo.Update(ctx, attempt.ID.Hex(), UpdateAuthAttemptDTO{
			Success:     false,
			AttemptLeft: &attemptsLeft,
		})
		if err != nil {
			s.log.Sugar.Errorf("Failed to decrement auth attempt left field for '%s': %v", attempt.ID, err)
			return err
		}
	}

	return ErrInvalidCode
}

// generateTokenPair generates a new access and refresh JWT token pair for the given user.
func (s *AuthServiceImpl) generateTokenPair(
	user *users.ReadUserDTO,
) (*TokenPairDTO, error) {
	accessToken, err := utils.CreateToken(
		s.cfg.AccessTokenSecret,
		user.ID,
		s.cfg.AccessTokenExpires,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := utils.CreateToken(
		s.cfg.RefreshTokenSecret,
		user.ID,
		s.cfg.RefreshTokenExpires,
	)
	if err != nil {
		return nil, err
	}

	return &TokenPairDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
