package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/common"
	"github.com/brokeboycoding/tempo/pkg/crypto"
)

type SecretHandler struct {
	secretRepo    storage.SecretRepository
	cfg           *config.Config
	logger        *logrus.Logger
	encryptionKey string
}

func NewSecretHandler(secretRepo storage.SecretRepository, cfg *config.Config) *SecretHandler {
	return &SecretHandler{
		secretRepo:    secretRepo,
		cfg:           cfg,
		logger:        common.GetLogger(),
		encryptionKey: cfg.EncryptionKey,
	}
}

type CreateSecretRequest struct {
	Name  string                 `json:"name" binding:"required"`
	Type  string                 `json:"type" binding:"required"`  // generic type hint (e.g., api_key, oauth_token)
	Value map[string]interface{} `json:"value" binding:"required"` // arbitrary key/values to store
}

type UpdateSecretRequest struct {
	Name  *string                `json:"name"`
	Type  *string                `json:"type"`
	Value map[string]interface{} `json:"value"`
}

type SecretResponse struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Value     map[string]interface{} `json:"value,omitempty"` // only present when explicitly requested
}

func (h *SecretHandler) CreateSecret(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateSecretRequest
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	if req.Name == "" || req.Type == "" || req.Value == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, type, and value are required"})
		return
	}

	// marshal the value and encrypt
	plain, err := json.Marshal(req.Value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value payload"})
		return
	}
	encrypted, err := crypto.Encrypt(string(plain), h.encryptionKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt secret"})
		return
	}

	secret := &domain.Secret{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      req.Name,
		Type:      req.Type,
		Value:     encrypted,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := h.secretRepo.Create(ctx, secret); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create secret"})
		return
	}

	c.JSON(http.StatusCreated, SecretResponse{
		ID:        secret.ID,
		UserID:    secret.UserID,
		Name:      secret.Name,
		Type:      secret.Type,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
	})
}

func (h *SecretHandler) ListSecrets(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	secrets, err := h.secretRepo.ListByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list secrets"})
		return
	}
	total, err := h.secretRepo.GetCount(ctx, userID)
	if err != nil {
		total = 0
	}

	resp := make([]SecretResponse, len(secrets))
	for i, s := range secrets {
		resp[i] = SecretResponse{
			ID:        s.ID,
			UserID:    s.UserID,
			Name:      s.Name,
			Type:      s.Type,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"secrets":   resp,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *SecretHandler) GetSecret(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	//pragma: allowlist secret
	secret, err := h.secretRepo.GetByID(ctx, id)

	if err != nil || secret == nil || secret.UserID != userID { //pragma: allowlist secret
		c.JSON(http.StatusNotFound, gin.H{"error": "secret not found"})
		return
	}

	plain, err := crypto.Decrypt(secret.Value, h.encryptionKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt secret"})
		return
	}

	var val map[string]interface{}
	_ = json.Unmarshal([]byte(plain), &val)
	c.JSON(http.StatusOK, SecretResponse{
		ID:        secret.ID,
		UserID:    secret.UserID,
		Name:      secret.Name,
		Type:      secret.Type,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
		Value:     val,
	})
}

func (h *SecretHandler) UpdateSecret(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	var req UpdateSecretRequest
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	secret, err := h.secretRepo.GetByID(ctx, id)
	if err != nil || secret == nil || secret.UserID != userID { //pragma: allowlist secret
		c.JSON(http.StatusNotFound, gin.H{"error": "secret not found"})
		return
	}

	if req.Name != nil {
		secret.Name = *req.Name
	}
	if req.Type != nil {
		secret.Type = *req.Type
	}
	if req.Value != nil {
		plain, err := json.Marshal(req.Value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value payload"})
			return
		}
		encrypted, err := crypto.Encrypt(string(plain), h.encryptionKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt secret"})
			return
		}
		secret.Value = encrypted
	}
	secret.UpdatedAt = time.Now()

	if err := h.secretRepo.Update(ctx, secret); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update secret"})
		return
	}

	c.JSON(http.StatusOK, SecretResponse{
		ID:        secret.ID,
		UserID:    secret.UserID,
		Name:      secret.Name,
		Type:      secret.Type,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
	})
}

func (h *SecretHandler) DeleteSecret(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	secret, err := h.secretRepo.GetByID(ctx, id)
	if err != nil || secret == nil || secret.UserID != userID { //pragma: allowlist secret
		c.JSON(http.StatusNotFound, gin.H{"error": "secret not found"})
		return
	}

	if err := h.secretRepo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete secret"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
