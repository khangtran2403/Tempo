package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {

	password := "mysecretpassword" //pragma: allowlist secret
	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestVerifyPassword(t *testing.T) {
	password := "mysecretpassword" //pragma: allowlist secret
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	// Test correct password
	assert.True(t, VerifyPassword(password, hash), "Correct password should verify successfully")

	// Test incorrect password
	assert.False(t, VerifyPassword("wrongpassword", hash), "Incorrect password should fail verification")
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	// Test with a string that is not a valid bcrypt hash
	assert.False(t, VerifyPassword("anypassword", "notavalidhash"), "Verification should fail for an invalid hash format")
}
