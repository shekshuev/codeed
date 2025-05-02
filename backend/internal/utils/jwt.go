package utils

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// ContextKey is a custom type for context keys used in JWT middleware.
type ContextKey string

const (
	// AccessTokenCookieName is the name of the cookie where access token is stored.
	AccessTokenCookieName = "X-Access-Token"
	// RefreshTokenCookieName is the name of the cookie where refresh token is stored.
	RefreshTokenCookieName = "X-Refresh-Token"
	// ContextClaimsKey is the context key used to store JWT claims.
	ContextClaimsKey = ContextKey("user-claims")
	// JWTIssuer used in 'iss' field of JWT
	JWTIssuer = "Code | Ed"
)

// Predefined errors related to JWT token parsing and validation.
var (
	ErrInvalidSignature = errors.New("token signature is invalid")
	ErrTokenExpired     = errors.New("token is expired")
	ErrTokenInvalid     = errors.New("token is invalid")
)

// GetRawAccessToken extracts the access token from the Authorization header (Bearer <token> format).
func GetRawAccessToken(req *http.Request) (string, error) {
	authHeader := req.Header.Get("Authorization")
	if len(authHeader) == 0 || !strings.HasPrefix(authHeader, "Bearer ") {
		return "", ErrTokenInvalid
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

// GetRawRefreshToken extracts the refresh token from cookies using the predefined cookie name.
func GetRawRefreshToken(req *http.Request) (string, error) {
	cookie, err := req.Cookie(RefreshTokenCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// CreateToken creates a new JWT string using the provided secret, user ID and expiration duration.
// The token contains standard registered claims including Issuer, Subject, ExpiresAt, IssuedAt, and ID.
func CreateToken(secret, userId string, exp time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    JWTIssuer,
		Subject:   userId,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        uuid.New().String(),
	})
	return token.SignedString([]byte(secret))
}

// ParseToken parses the given JWT string and validates its signature and expiration time.
// Returns the extracted claims if valid, or an appropriate error if not.
func ParseToken(tokenString, secret string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		if validationErr, ok := err.(*jwt.ValidationError); ok {
			if validationErr.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrTokenExpired
			}
			if validationErr.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				return nil, ErrInvalidSignature
			}
		}
		return nil, ErrTokenInvalid
	}

	if !token.Valid {
		return nil, ErrInvalidSignature
	}
	return claims, nil
}

// GetClaimsFromContext retrieves JWT claims from the context, if present.
func GetClaimsFromContext(ctx context.Context) (jwt.RegisteredClaims, bool) {
	claims, ok := ctx.Value(ContextClaimsKey).(jwt.RegisteredClaims)
	return claims, ok
}

// PutClaimsToContext attaches JWT claims to the provided context.
func PutClaimsToContext(ctx context.Context, claims jwt.RegisteredClaims) context.Context {
	return context.WithValue(ctx, ContextClaimsKey, claims)
}
