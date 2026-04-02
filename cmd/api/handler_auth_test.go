package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brokeboycoding/tempo/internal/auth"
	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupTestRouter initializes a router for testing the AuthHandler
func setupAuthTestRouter() (*gin.Engine, *MockUserRepository) {
	router := gin.New()
	mockUserRepo := new(MockUserRepository)
	cfg := &config.Config{
		JWT: struct{ Secret string }{Secret: "test-secret-for-api"},
	}
	authHandler := NewAuthHandler(mockUserRepo, cfg)

	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	return router, mockUserRepo
}

func TestRegister(t *testing.T) {
	router, mockUserRepo := setupAuthTestRouter()
	t.Run("Successful Registration", func(t *testing.T) {
		// Setup mock
		mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil).Once()
		mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Once()
		// Prepare request
		reqBody := RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)
		var respBody AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.NotEmpty(t, respBody.Token)
		assert.Equal(t, "test@example.com", respBody.User.Email)
		assert.Equal(t, "Test User", respBody.User.Name)
	})
	t.Run("Email Already Registered", func(t *testing.T) {
		// Setup mock for existing user
		existingUser := &domain.User{ID: "1", Email: "existing@example.com"}
		mockUserRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil).Once()
		// Prepare request
		reqBody := RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
			Name:     "Another User",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}
func TestLogin(t *testing.T) {
	router, mockUserRepo := setupAuthTestRouter()
	hashedPassword, _ := auth.HashPassword("password123")
	user := &domain.User{
		ID:           "user-1",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Name:         "Test User",
		IsActive:     true,
		CreatedAt:    time.Now(),
	}
	t.Run("Successful Login", func(t *testing.T) {
		// Setup mock
		mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil).Once()
		// Prepare request
		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var respBody AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.NotEmpty(t, respBody.Token)
		assert.Equal(t, "test@example.com", respBody.User.Email)
	})
	t.Run("Invalid Password", func(t *testing.T) {
		mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil).Once()
		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("User Not Found", func(t *testing.T) {
		mockUserRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, nil).Once()
		reqBody := LoginRequest{
			Email:    "notfound@example.com",
			Password: "anypassword",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
