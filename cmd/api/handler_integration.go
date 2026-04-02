package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	gcs "cloud.google.com/go/storage"
	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/common"
	"github.com/brokeboycoding/tempo/pkg/crypto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

// ... (rest of the file)

type IntegrationHandler struct {
	integrationRepo   storage.IntegrationRepository
	cfg               *config.Config
	logger            *logrus.Logger
	googleOAuthConfig *oauth2.Config
	githubOAuthConfig *oauth2.Config
	notionOAuthConfig *oauth2.Config
	encryptionKey     string
}

type GoogleUserInfo1 struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func NewIntegrationHandler(integrationRepo storage.IntegrationRepository, cfg *config.Config) *IntegrationHandler {
	googleOAuthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleConnect.ClientID,
		ClientSecret: cfg.GoogleConnect.ClientSecret,
		RedirectURL:  cfg.OAuth.GoogleIntegrationRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			drive.DriveFileScope,
			sheets.SpreadsheetsScope,
			gcs.ScopeReadWrite,
		},
		Endpoint: google.Endpoint,
	}

	githubOAuthConfig := &oauth2.Config{
		ClientID:     cfg.GitHub.ClientID,
		ClientSecret: cfg.GitHub.ClientSecret,
		RedirectURL:  cfg.OAuth.GitHubIntegrationRedirectURL,
		Scopes:       []string{"repo", "user:email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}

	notionOAuthConfig := &oauth2.Config{
		ClientID:     cfg.Notion.ClientID,
		ClientSecret: cfg.Notion.ClientSecret,
		RedirectURL:  cfg.OAuth.NotionIntegrationRedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.notion.com/v1/oauth/authorize",
			TokenURL: "https://api.notion.com/v1/oauth/token",
		},
	}

	return &IntegrationHandler{
		integrationRepo:   integrationRepo,
		cfg:               cfg,
		logger:            common.GetLogger(),
		googleOAuthConfig: googleOAuthConfig,
		githubOAuthConfig: githubOAuthConfig,
		notionOAuthConfig: notionOAuthConfig,
		encryptionKey:     cfg.EncryptionKey,
	}
}

func (h *IntegrationHandler) getFrontendURL() string {
	return "http://localhost:3000"
}

// --- Google Handlers ---

func (h *IntegrationHandler) HandleGoogleConnect(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	state := userID
	c.SetCookie("oauthstate_integration", state, 3600, "/", c.Request.URL.Hostname(), false, true)
	url := h.googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *IntegrationHandler) HandleGoogleCallback(c *gin.Context) {
	userID, err := c.Cookie("oauthstate_integration")
	if err != nil || c.Query("state") != userID {
		h.logger.Errorf("Invalid oauth state for google integration: cookie_err=%v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=invalid_state")
		return
	}

	code := c.Query("code")
	token, err := h.googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		h.logger.Errorf("Failed to exchange code for google integration: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=code_exchange_failed")
		return
	}

	client := h.googleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		h.logger.Errorf("Failed to get google user info: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=userinfo_failed")
		return
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		h.logger.Errorf("Failed to decode google user info: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=userinfo_decode_failed")
		return
	}

	h.saveIntegration(c, "google", userInfo.Email, userID, token)
}

// --- GitHub Handlers ---

func (h *IntegrationHandler) HandleGitHubConnect(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	state := userID
	c.SetCookie("oauthstate_integration", state, 3600, "/", c.Request.URL.Hostname(), false, true)
	url := h.githubOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *IntegrationHandler) HandleGitHubCallback(c *gin.Context) {
	userID, err := c.Cookie("oauthstate_integration")
	if err != nil || c.Query("state") != userID {
		h.logger.Errorf("Invalid oauth state for github integration: cookie_err=%v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=invalid_state")
		return
	}

	code := c.Query("code")
	token, err := h.githubOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		h.logger.Errorf("Failed to exchange code for github integration: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=code_exchange_failed")
		return
	}

	client := h.githubOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		h.logger.Errorf("Failed to get github user emails: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=userinfo_failed")
		return
	}
	defer resp.Body.Close()

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		h.logger.Errorf("Failed to decode github user emails: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=userinfo_decode_failed")
		return
	}

	primaryEmail := fmt.Sprintf("GitHub Account (%s)", userID)
	for _, email := range emails {
		if email.Primary {
			primaryEmail = email.Email
			break
		}
	}

	h.saveIntegration(c, "github", primaryEmail, userID, token)
}

// --- Notion Handlers ---

func (h *IntegrationHandler) HandleNotionConnect(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	state := userID
	c.SetCookie("oauthstate_integration", state, 3600, "/", c.Request.URL.Hostname(), false, true)
	url := h.notionOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *IntegrationHandler) HandleNotionCallback(c *gin.Context) {
	userID, err := c.Cookie("oauthstate_integration")
	if err != nil || c.Query("state") != userID {
		h.logger.Errorf("Invalid oauth state for notion integration: cookie_err=%v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=invalid_state")
		return
	}

	code := c.Query("code")
	token, err := h.notionOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		h.logger.Errorf("Failed to exchange code for notion integration: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=code_exchange_failed")
		return
	}

	workspaceName := "Notion Workspace"
	if extra, ok := token.Extra("workspace_name").(string); ok {
		workspaceName = extra
	}

	h.saveIntegration(c, "notion", workspaceName, userID, token)
}

// --- Common & CRUD Handlers ---

func (h *IntegrationHandler) saveIntegration(c *gin.Context, provider, accountName, userID string, token *oauth2.Token) {
	tokenJson, err := json.Marshal(token)
	if err != nil {
		h.logger.Errorf("Failed to marshal token for %s: %v", provider, err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=token_serialization_failed")
		return
	}

	encryptedToken, err := crypto.Encrypt(string(tokenJson), h.encryptionKey)
	if err != nil {
		h.logger.Errorf("Failed to encrypt token for %s: %v", provider, err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=token_encryption_failed")
		return
	}

	integration := &domain.Integration{
		ID:          uuid.New().String(),
		UserID:      userID,
		Provider:    provider,
		Name:        accountName,
		AccessToken: encryptedToken,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.integrationRepo.Create(c.Request.Context(), integration); err != nil {
		h.logger.Errorf("Failed to save %s integration: %v", provider, err)
		c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?error=database_error")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, h.getFrontendURL()+"/integrations?success=true")
}

func (h *IntegrationHandler) ListIntegrations(c *gin.Context) {
	userID := GetUserID(c)
	integrations, err := h.integrationRepo.ListByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list integrations"})
		return
	}
	if integrations == nil {
		integrations = []*domain.Integration{} // Return empty list instead of null
	}
	c.JSON(http.StatusOK, gin.H{"integrations": integrations})
}

func (h *IntegrationHandler) DeleteIntegration(c *gin.Context) {
	userID := GetUserID(c)
	integrationID := c.Param("id")

	integration, err := h.integrationRepo.GetByID(c.Request.Context(), integrationID)
	if err != nil || integration == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}

	if integration.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.integrationRepo.Delete(c.Request.Context(), integrationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete integration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
