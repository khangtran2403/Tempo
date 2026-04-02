package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID               string `json:"user_id"`
	Email                string `json:"email"`
	jwt.RegisteredClaims        // Standard claims: exp, iat, iss, etc
}

func GenerateToken(userID, email, secret string) (string, error) {
	claims := &TokenClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			// Token hết hạn sau 24 giờ
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			// Thời gian phát hành
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", nil
	}
	return tokenString, nil
}

// VerifyToken xác minh JWT token
func VerifyToken(tokenString, secret string) (*TokenClaims, error) {
	// Parse token với custom claims
	token, err := jwt.ParseWithClaims(
		tokenString,
		&TokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Kiểm tra signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}
	// Lấy claims
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Kiểm tra expiration (GORM tự động check, nhưng cũng kiểm tra lại)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}
