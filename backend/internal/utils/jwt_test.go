package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateAndParseToken(t *testing.T) {
	secret := "test-secret"
	userID := "user-123"

	t.Run("valid token is parsed correctly", func(t *testing.T) {
		tokenStr, err := CreateToken(secret, userID, time.Minute)
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenStr)

		claims, err := ParseToken(tokenStr, secret)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.Subject)
		assert.Equal(t, JWTIssuer, claims.Issuer)
	})

	t.Run("expired token returns error", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Issuer:    JWTIssuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
		})
		tokenStr, err := token.SignedString([]byte(secret))
		assert.NoError(t, err)

		_, err = ParseToken(tokenStr, secret)
		assert.ErrorIs(t, err, ErrTokenExpired)
	})

	t.Run("invalid signature returns error", func(t *testing.T) {
		tokenStr, err := CreateToken(secret, userID, time.Minute)
		assert.NoError(t, err)

		_, err = ParseToken(tokenStr, "wrong-secret")
		assert.ErrorIs(t, err, ErrInvalidSignature)
	})

	t.Run("invalid algorithm returns error", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
			Subject: userID,
		})
		tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		assert.NoError(t, err)

		_, err = ParseToken(tokenStr, "")
		assert.ErrorIs(t, err, ErrTokenInvalid)
	})
}

func TestGetRawAccessToken(t *testing.T) {
	t.Run("returns token from valid header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		token, err := GetRawAccessToken(req)
		assert.NoError(t, err)
		assert.Equal(t, "test-token", token)
	})

	t.Run("returns error if header is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		_, err := GetRawAccessToken(req)
		assert.ErrorIs(t, err, ErrTokenInvalid)
	})

	t.Run("returns error if header prefix is invalid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Token test")
		_, err := GetRawAccessToken(req)
		assert.ErrorIs(t, err, ErrTokenInvalid)
	})
}

func TestGetRawRefreshToken(t *testing.T) {
	t.Run("returns token from cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  RefreshTokenCookieName,
			Value: "refresh-token",
		})

		token, err := GetRawRefreshToken(req)
		assert.NoError(t, err)
		assert.Equal(t, "refresh-token", token)
	})

	t.Run("returns error if cookie is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		_, err := GetRawRefreshToken(req)
		assert.Error(t, err)
	})
}

func TestClaimsContext(t *testing.T) {
	t.Run("claims round-trip through context", func(t *testing.T) {
		ctx := context.Background()
		claims := jwt.RegisteredClaims{Subject: "user-999"}
		newCtx := PutClaimsToContext(ctx, claims)

		got, ok := GetClaimsFromContext(newCtx)
		assert.True(t, ok)
		assert.Equal(t, "user-999", got.Subject)
	})

	t.Run("GetClaimsFromContext returns false if not set", func(t *testing.T) {
		ctx := context.Background()
		_, ok := GetClaimsFromContext(ctx)
		assert.False(t, ok)
	})
}
