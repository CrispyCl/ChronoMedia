package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrExpiredToken   = errors.New("token has expired")
	ErrInvalidToken   = errors.New("token signature is invalid")
	ErrEmptySecretKey = errors.New("jwt secret key cannot be empty")
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secretKey []byte
}

func NewTokenManager(secret string) (*TokenManager, error) {
	if secret == "" {
		return nil, ErrEmptySecretKey
	}
	return &TokenManager{secretKey: []byte(secret)}, nil
}

func (m *TokenManager) GenerateJWT(userID, email, role string, ttl time.Duration) (string, error) {
	const op = "jwt.TokenManager.GenerateJWT"

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("%s: failed to sign token payload: %w", op, err)
	}

	return signed, nil
}

func (m *TokenManager) ValidateJWT(tokenString string) (*Claims, error) {
	const op = "jwt.TokenManager.ValidateJWT"

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s: unexpected signing method: %v", op, t.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("%s: %w", op, ErrExpiredToken)
		}
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	return claims, nil
}

func GenerateRandomToken(n int) (string, error) {
	const op = "jwt.GenerateRandomToken"

	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("%s: failed to read secure random bytes: %w", op, err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
