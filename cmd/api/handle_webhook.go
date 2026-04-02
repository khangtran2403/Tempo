package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/internal/workflow"
	"github.com/brokeboycoding/tempo/pkg/common"
)

type WebhookHandler struct {
	workflowRepo       storage.WorkflowRepository
	WebhookHistoryRepo storage.WebhookHistoryRepository
	temporalClient     client.Client
	cfg                *config.Config
	logger             *logrus.Logger
}

func NewWebhookHandler(
	workflowRepo storage.WorkflowRepository,
	WebhookHistoryRepo storage.WebhookHistoryRepository,
	temporalClient client.Client,
	cfg *config.Config,
) *WebhookHandler {
	return &WebhookHandler{
		workflowRepo:       workflowRepo,
		WebhookHistoryRepo: WebhookHistoryRepo,
		temporalClient:     temporalClient,
		cfg:                cfg,
		logger:             common.GetLogger(),
	}
}

func (h *WebhookHandler) TriggerWebhook(c *gin.Context) {
	workflowID := c.Param("workflow_id")

	h.logger.Infof("Webhook triggered for workflow: %s", workflowID)

	wf, err := h.workflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil {
		h.logger.Errorf("Error getting workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal error",
		})
		return
	}

	if wf == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	if !wf.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workflow is not active",
		})
		return
	}

	expectedToken := h.getWebhookToken(wf)
	providedToken := c.Query("token")

	if expectedToken != "" && providedToken != expectedToken {
		h.logger.Warnf("Invalid webhook token for workflow %s", workflowID)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})
		return
	}

	if signature := c.GetHeader("X-Webhook-Signature"); signature != "" {
		if !h.validateSignature(c, signature, wf) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid signature",
			})
			return
		}
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.logger.Warnf("Webhook for workflow %s received a non-JSON or empty body. Proceeding with empty payload. Error: %v", workflowID, err)
		// Initialize payload as an empty map if binding fails
		payload = make(map[string]interface{})
	}

	payload["_metadata"] = map[string]interface{}{
		"received_at": time.Now().Format(time.RFC3339),
		"headers":     h.extractHeaders(c),
		"method":      c.Request.Method,
		"url":         c.Request.URL.String(),
		"ip":          c.ClientIP(),
	}

	executionID := uuid.New().String()

	workflowOptions := client.StartWorkflowOptions{
		ID:                       executionID,
		TaskQueue:                "workflow-queue",
		WorkflowExecutionTimeout: 30 * time.Minute,
	}

	req := &workflow.ExecuteWorkflowRequest{
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		UserID:      wf.UserID,
		TriggerData: payload,
		Definition:  wf.Definition,
	}

	execution, err := h.temporalClient.ExecuteWorkflow(
		c.Request.Context(),
		workflowOptions,
		workflow.ExecuteWorkflowWorkflow,
		req,
	)

	if err != nil {
		h.logger.Errorf("Failed to start workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to start workflow",
		})
		return
	}

	h.logger.Infof("✅ Workflow started: execution_id=%s, temporal_id=%s",
		executionID, execution.GetID())

	c.JSON(http.StatusAccepted, gin.H{
		"success":      true,
		"execution_id": executionID,
		"workflow_id":  execution.GetID(),
		"run_id":       execution.GetRunID(),
		"status":       "processing",
		"message":      "Workflow execution started",
	})
	historyID := uuid.New().String()
	history := &domain.WebhookHistory{
		ID:         historyID,
		WorkflowID: workflowID,
		Method:     c.Request.Method,
		Headers:    h.extractHeaders(c),
		Body:       payload,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		ReceivedAt: time.Now(),
	}
	if err != nil {

		history.Status = "failed"
		history.Error = err.Error()
	} else {
		history.Status = "success"
		history.ExecutionID = executionID
	}

	go func() {
		ctx := context.Background()
		if err := h.WebhookHistoryRepo.Create(ctx, history); err != nil {
			h.logger.Errorf("Failed to save webhook history: %v", err)
		}
	}()
}

// getWebhookToken lấy webhook token từ workflow config
func (h *WebhookHandler) getWebhookToken(wf *domain.Workflow) string {
	// Token được lưu trong trigger config
	if wf.Definition.Trigger.Config != nil {
		if token, ok := wf.Definition.Trigger.Config["webhook_token"].(string); ok {
			return token
		}
	}
	return ""
}

func (h *WebhookHandler) validateSignature(
	c *gin.Context,
	providedSignature string,
	wf *domain.Workflow,
) bool {
	// Get webhook secret từ config
	secret := ""
	if wf.Definition.Trigger.Config != nil {
		if s, ok := wf.Definition.Trigger.Config["webhook_secret"].(string); ok {
			secret = s
		}
	}
	//pragma: allowlist secret
	if secret == "" {
		// Nếu không có secret, bỏ qua validation
		return true //pragma: allowlist secret
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(bodyBytes)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(providedSignature))
}

func (h *WebhookHandler) extractHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	relevantHeaders := []string{
		"Content-Type",
		"User-Agent",
		"X-GitHub-Event",
		"X-Stripe-Signature",
		"X-Hub-Signature",
	}

	for _, key := range relevantHeaders {
		if value := c.GetHeader(key); value != "" {
			headers[key] = value
		}
	}

	return headers
}
func (h *WebhookHandler) ListWebhookHistory(c *gin.Context) {
	userID := GetUserID(c)
	workflowID := c.Param("id")

	wf, err := h.workflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil || wf == nil || wf.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	history, err := h.WebhookHistoryRepo.ListByWorkflowID(
		c.Request.Context(),
		workflowID,
		pageSize,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get webhook history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history":   history,
		"page":      page,
		"page_size": pageSize,
	})
}

// Replay một webhook call từ history
func (h *WebhookHandler) ReplayWebhook(c *gin.Context) {
	userID := GetUserID(c)
	workflowID := c.Param("id")
	historyID := c.Param("history_id")

	wf, err := h.workflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil || wf == nil || wf.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	history, err := h.WebhookHistoryRepo.GetByID(c.Request.Context(), historyID)
	if err != nil || history == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "webhook history not found",
		})
		return
	}

	if history.WorkflowID != workflowID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	// REPLAY WEBHOOK
	executionID := uuid.New().String()
	payload := history.Body
	if payload == nil {
		payload = make(map[string]interface{})
	}
	payload["_replay"] = true
	payload["_original_webhook_id"] = historyID
	payload["_replayed_at"] = time.Now()

	workflowOptions := client.StartWorkflowOptions{
		ID:        executionID,
		TaskQueue: "workflow-queue",
	}

	req := &workflow.ExecuteWorkflowRequest{
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		UserID:      wf.UserID,
		TriggerData: payload,
		Definition:  wf.Definition,
	}

	execution, err := h.temporalClient.ExecuteWorkflow(
		c.Request.Context(),
		workflowOptions,
		workflow.ExecuteWorkflowWorkflow,
		req,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to replay webhook",
		})
		return
	}

	h.logger.Infof("✅ Webhook replayed: history_id=%s, execution_id=%s",
		historyID, executionID)

	c.JSON(http.StatusAccepted, gin.H{
		"success":      true,
		"execution_id": executionID,
		"workflow_id":  execution.GetID(),
		"message":      "Webhook replayed successfully",
	})
}
