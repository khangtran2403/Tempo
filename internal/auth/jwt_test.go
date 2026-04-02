package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAndVerifyToken(t *testing.T) {

	secret := "test-secret-key-that-is-long-enough" //pragma: allowlist secret
	userID := "user-123"
	email := "test@example.com"

	// Generate token
	tokenString, err := GenerateToken(userID, email, secret)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Verify valid token
	claims, err := VerifyToken(tokenString, secret)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), claims.ExpiresAt.Time, 2*time.Second)
}

// pragma: allowlist secret
func TestVerifyToken_InvalidSignature(t *testing.T) {

	secret := "correct-secret"    //pragma: allowlist secret
	wrongSecret := "wrong-secret" //pragma: allowlist secret
	userID := "user-123"
	email := "test@example.com"

	tokenString, err := GenerateToken(userID, email, secret)
	assert.NoError(t, err)

	// Try to verify with the wrong secret
	claims, err := VerifyToken(tokenString, wrongSecret)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, jwt.ErrSignatureInvalid)
}

func TestVerifyToken_Expired(t *testing.T) {

	secret := "test-secret" //pragma: allowlist secret
	userID := "user-123"
	email := "test@example.com"

	// Create an already expired token
	claims := &TokenClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredTokenString, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)

	// Try to verify the expired token
	verifiedClaims, err := VerifyToken(expiredTokenString, secret)
	assert.Error(t, err)
	assert.Nil(t, verifiedClaims)
	assert.ErrorIs(t, err, jwt.ErrTokenExpired)
}

func TestVerifyToken_InvalidFormat(t *testing.T) {

	secret := "any-secret" //pragma: allowlist secret

	// Test with a malformed token string
	claims, err := VerifyToken("this.is.not.a.jwt", secret)
	assert.Error(t, err)
	assert.Nil(t, claims)
}
