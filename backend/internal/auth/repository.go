package auth

import (
	"context"
)

// AuthAttemptRepository defines methods to manage auth attempts in the database.
type AuthAttemptRepository interface {
	Create(ctx context.Context, attempt AuthAttempt) (*AuthAttempt, error)
	GetByID(ctx context.Context, id string) (*AuthAttempt, error)
	Update(ctx context.Context, id string, dto UpdateAuthAttemptDTO) error
	Delete(ctx context.Context, id string) error
}
