package auth

import (
	"time"

	"github.com/shekshuev/codeed/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthType represents the type of authentication used.
type AuthType int

const (
	// TypeTelegram indicates authentication via Telegram username.
	TypeTelegram AuthType = iota
	// TypeEmail indicates authentication via email.
	TypeEmail
)

const (
	// AttemptTTL defines how long a code is valid after generation.
	AttemptTTL = 5 * time.Minute

	// MaxAttempts defines how many tries a user has before lockout.
	MaxAttempts = 3

	// CodeLength defines the number of digits in a generated code.
	CodeLength = 6
)

// AuthAttempt represents a single authorization attempt.
// Used to store and track code-based login requests.
type AuthAttempt struct {
	ID             primitive.ObjectID `bson:"_id"`
	IdentifierUsed string             `bson:"identifier_used"`
	Type           AuthType           `bson:"type"`
	Code           string             `bson:"code"`
	Success        bool               `bson:"success"`
	AttemptLeft    int                `bson:"attempt_left"`
	TTL            time.Duration      `bson:"ttl"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}

// CreateTelegramCodeRequestDTO represents an incoming request to start Telegram login flow.
type CreateTelegramCodeRequestDTO struct {
	TelegramUsername string `json:"telegram_username" validate:"required"`
}

// ReadTelegramCodeRequestDTO is returned to the user after requesting a login code.
// Includes a wait timeout to prevent rapid retries.
type ReadTelegramCodeRequestDTO struct {
	TelegramUsername string    `json:"telegram_username"`
	ID               string    `json:"id"`
	WaitUntil        time.Time `json:"wait_until"`
}

// CreateCheckTelegramCodeRequestDTO is sent by the client to verify the received code.
type CreateCheckTelegramCodeRequestDTO struct {
	ID               string `json:"id" validate:"required"`
	TelegramUsername string `json:"telegram_username" validate:"required"`
	Code             string `json:"code" validate:"required"`
}

// TokenPairDTO contains an access and refresh token issued after successful login.
type TokenPairDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ReadAuthAttemptDTO is a read-only representation of an AuthAttempt for external usage (e.g. API responses).
type ReadAuthAttemptDTO struct {
	ID             string        `json:"id"`
	IdentifierUsed string        `json:"identifier_used"`
	Type           AuthType      `json:"type"`
	Code           string        `json:"code"`
	Success        bool          `json:"success"`
	AttemptLeft    int           `json:"attempt_left"`
	TTL            time.Duration `json:"ttl"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// UpdateAuthAttemptDTO is used to update an AuthAttempt in the database.
type UpdateAuthAttemptDTO struct {
	Success     *bool
	AttemptLeft *int
}

// ToReadTelegramCodeRequestDTO converts an AuthAttempt into a DTO sent after code generation.
func (u AuthAttempt) ToReadTelegramCodeRequestDTO() *ReadTelegramCodeRequestDTO {
	return &ReadTelegramCodeRequestDTO{
		ID:               u.ID.Hex(),
		TelegramUsername: u.IdentifierUsed,
		WaitUntil:        u.CreatedAt.Add(u.TTL),
	}
}

// ToReadAuthAttemptDTO maps the AuthAttempt to a read-only DTO, e.g. for admin panels or logs.
func (u AuthAttempt) ToReadAuthAttemptDTO() *ReadAuthAttemptDTO {
	return &ReadAuthAttemptDTO{
		ID:             u.ID.Hex(),
		IdentifierUsed: u.IdentifierUsed,
		Type:           u.Type,
		Code:           u.Code,
		Success:        u.Success,
		AttemptLeft:    u.AttemptLeft,
		TTL:            u.TTL,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

// ToAuthAttemptFromCreateTelegramCodeRequestDTO creates a new AuthAttempt based on a user's Telegram login request.
func (dto CreateTelegramCodeRequestDTO) ToAuthAttemptFromCreateTelegramCodeRequestDTO() AuthAttempt {
	return AuthAttempt{
		ID:             primitive.NewObjectID(),
		IdentifierUsed: dto.TelegramUsername,
		Type:           TypeTelegram,
		Code:           utils.RandomDigits(CodeLength),
		Success:        false,
		TTL:            AttemptTTL,
		AttemptLeft:    MaxAttempts,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
}
