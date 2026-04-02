package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/brokeboycoding/tempo/internal/auth"
	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthHandler struct {
	UserRepo          storage.UserRepository
	cfg               *config.Config
	logger            *logrus.Logger
	googleOAuthConfig *oauth2.Config
}

func NewAuthHandler(UserRepo storage.UserRepository, cfg *config.Config) *AuthHandler {
	googleOAuthConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.OAuth.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &AuthHandler{
		UserRepo:          UserRepo,
		cfg:               cfg,
		logger:            common.GetLogger(),
		googleOAuthConfig: googleOAuthConfig,
	}
}

type GoogleUserInfo struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func (h *AuthHandler) HandleGoogleLogin(c *gin.Context) {
	state := uuid.New().String()
	c.SetCookie("oauthstate", state, 3600, "/", c.Request.URL.Hostname(), false, true)

	url := h.googleOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) HandleGoogleCallback(c *gin.Context) {
	oauthState, err := c.Cookie("oauthstate")
	if err != nil || c.Query("state") != oauthState {
		h.logger.Errorf("Invalid oauth state: cookie=%s, query=%s", oauthState, c.Query("state"))
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=invalid_state", h.getFrontendURL()))
		return
	}

	code := c.Query("code")
	token, err := h.googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		h.logger.Errorf("Failed to exchange code: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=code_exchange_failed", h.getFrontendURL()))
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		h.logger.Errorf("Failed to get user info: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=userinfo_failed", h.getFrontendURL()))
		return
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		h.logger.Errorf("Failed to read user info body: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=userinfo_read_failed", h.getFrontendURL()))
		return
	}

	var userInfo GoogleUserInfo
	json.Unmarshal(contents, &userInfo)

	if !userInfo.VerifiedEmail {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=email_not_verified", h.getFrontendURL()))
		return
	}

	user, err := h.UserRepo.GetByEmail(c.Request.Context(), userInfo.Email)
	if err != nil {
		h.logger.Errorf("Error checking for existing user: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=database_error", h.getFrontendURL()))
		return
	}

	if user == nil { // User does not exist, create new one
		user = &domain.User{
			ID:        uuid.New().String(),
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := h.UserRepo.Create(c.Request.Context(), user); err != nil {
			h.logger.Errorf("Failed to create user from Google auth: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=user_creation_failed", h.getFrontendURL()))
			return
		}
		h.logger.Infof("New user created via Google OAuth: %s", user.Email)
	}

	// 5. Generate local JWT
	jwtToken, err := auth.GenerateToken(user.ID, user.Email, h.cfg.JWT.Secret)
	if err != nil {
		h.logger.Errorf("Failed to generate JWT for OAuth user: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/login?error=token_generation_failed", h.getFrontendURL()))
		return
	}

	// 6. Redirect back to frontend with token
	userJSON, _ := json.Marshal(UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	})

	redirectURL := fmt.Sprintf(
		"%s/auth/google/callback?token=%s&user=%s",
		h.getFrontendURL(),
		jwtToken,
		url.QueryEscape(string(userJSON)),
	)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *AuthHandler) getFrontendURL() string {
	// In a real app, this should come from config
	return "http://localhost:3000"
}

// Request/Response Models
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	//kiem tra JSON format
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request format",
			"details": err.Error(),
		})
		return
	}

	// kiem tra email da ton tai chua
	existingUser, err := h.UserRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		h.logger.Errorf("Error checking existing user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "email already registered",
		})
		return
	}
	// Hash Password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.logger.Errorf("Error hashing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	//tao user object
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		Name:         req.Name,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	//Luu vao database
	if err := h.UserRepo.Create(c.Request.Context(), user); err != nil {
		h.logger.Errorf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
		return
	}
	// tao JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, h.cfg.JWT.Secret)
	if err != nil {
		h.logger.Errorf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}
	//Tra Response
	h.logger.Infof("New user registered: %s", user.Email)
	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User: &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
	})
}
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	// 1. Kiem tra JSON format
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request format",
			"details": err.Error(),
		})
		return
	}

	// 2. Get user by email
	user, err := h.UserRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		h.logger.Errorf("Error getting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// 3. Checks user co ton tai ko
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	// 4. Verify password
	if !auth.VerifyPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	// 5. Check if user hoat dong
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "account is inactive",
		})
		return
	}

	// 6. Tao JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, h.cfg.JWT.Secret)
	if err != nil {
		h.logger.Errorf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	h.logger.Infof("User logged in: %s", user.Email)

	// 7. Tra response
	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User: &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
	})
}
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Lấy user_id từ context (được set bởi AuthMiddleware)
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	// Lay user from database
	user, err := h.UserRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Errorf("Error getting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	})
}
