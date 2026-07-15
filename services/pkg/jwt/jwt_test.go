package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenManager(t *testing.T) {
	testCases := []struct {
		name    string
		secret  string
		wantErr error
	}{
		{
			name:    "valid secret",
			secret:  "test-secret-key",
			wantErr: nil,
		},
		{
			name:    "empty secret",
			secret:  "",
			wantErr: ErrEmptySecretKey,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewTokenManager(tc.secret)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestTokenManager_Lifecycle(t *testing.T) {
	const secret = "test-secret-key"
	manager, err := NewTokenManager(secret)
	require.NoError(t, err)

	t.Run("successful generate and validate", func(t *testing.T) {
		userID := "user-123"
		email := "user@example.com"
		role := "admin"
		ttl := 10 * time.Minute

		tokenStr, err := manager.GenerateJWT(userID, email, role, ttl)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenStr)

		claims, err := manager.ValidateJWT(tokenStr)
		require.NoError(t, err)
		require.NotNil(t, claims)

		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, role, claims.Role)
	})

	t.Run("expired token error", func(t *testing.T) {
		tokenStr, err := manager.GenerateJWT("user-123", "user@example.com", "user", -1*time.Minute)
		require.NoError(t, err)

		_, err = manager.ValidateJWT(tokenStr)
		assert.ErrorIs(t, err, ErrExpiredToken)
	})

	t.Run("invalid signature error", func(t *testing.T) {
		tokenStr, err := manager.GenerateJWT("user-123", "user@example.com", "user", time.Minute)
		require.NoError(t, err)

		wrongManager, err := NewTokenManager("wrong-test-secret-key")
		require.NoError(t, err)

		_, err = wrongManager.ValidateJWT(tokenStr)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("malformed token error", func(t *testing.T) {
		_, err := manager.ValidateJWT("invalid-token")
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}
