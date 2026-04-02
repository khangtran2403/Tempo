package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/internal/workflow"
	"github.com/brokeboycoding/tempo/pkg/common"

	"github.com/brokeboycoding/tempo/internal/connector"
	"github.com/brokeboycoding/tempo/internal/scheduler"

	"go.temporal.io/sdk/client"
)

type WorkflowHandler struct {
	WorkflowRepo   storage.WorkflowRepository
	versionRepo    storage.WorkflowVersionRepository
	auditLogger    storage.AuditLogRepository
	scheduler      scheduler.Scheduler
	temporalClient client.Client
	cfg            *config.Config
	logger         *logrus.Logger
}

func NewWorkflowHandler(WorkflowRepo storage.WorkflowRepository, versionRepo storage.WorkflowVersionRepository, auditLogger storage.AuditLogRepository, scheduler scheduler.Scheduler, temporalClient client.Client, cfg *config.Config) *WorkflowHandler {
	return &WorkflowHandler{
		WorkflowRepo:   WorkflowRepo,
		versionRepo:    versionRepo,
		auditLogger:    auditLogger,
		scheduler:      scheduler,
		temporalClient: temporalClient,
		cfg:            cfg,
		logger:         common.GetLogger(),
	}
}

type CreateWorkflowRequest struct {
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description"`
	Definition  domain.WorkflowDefinition `json:"definition" binding:"required"`
}

type UpdateWorkflowRequest struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Definition  *domain.WorkflowDefinition `json:"definition"`
	IsActive    *bool                      `json:"is_active"`
}

// Response Models
type WorkflowResponse struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Definition  domain.WorkflowDefinition `json:"definition"`
	Status      string                    `json:"status"`
	IsActive    bool                      `json:"is_active"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

type WorkflowListResponse struct {
	Workflows  []*WorkflowResponse `json:"workflows"`
	TotalCount int64               `json:"total_count"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}

// Lấy danh sách workflows của user (có pagination)
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	users := GetUserID(c)
	if users == "" {
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
	workflows, err := h.WorkflowRepo.ListByUserID(
		c.Request.Context(),
		users,
		pageSize,
		offset,
	)
	if err != nil {
		h.logger.Errorf("Error listing workflows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list workflows",
		})
		return
	}
	totalCount, err := h.WorkflowRepo.GetCount(c.Request.Context(), users)
	if err != nil {
		h.logger.Errorf("Error getting workflow count: %v", err)
		totalCount = 0
	}
	//tra ve response DTO cho client
	workflowResponses := make([]*WorkflowResponse, len(workflows))
	for i, wf := range workflows {
		workflowResponses[i] = &WorkflowResponse{
			ID:          wf.ID,
			Name:        wf.Name,
			Description: wf.Description,
			Definition:  wf.Definition,
			Status:      wf.Status,
			IsActive:    wf.IsActive,
			CreatedAt:   wf.CreatedAt,
			UpdatedAt:   wf.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, WorkflowListResponse{
		Workflows:  workflowResponses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	})
}
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request format",
			"details": err.Error(),
		})
		return
	}
	if err := h.ValidateWorkflowDefinition(&req.Definition); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid workflow definition",
			"details": err.Error(),
		})
		return
	}
	workflow := &domain.Workflow{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Definition:  req.Definition,
		Status:      "draft", // Default status
		IsActive:    false,   // Default inactive
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	//luu vao database
	if err := h.WorkflowRepo.Create(c.Request.Context(), workflow); err != nil {
		h.logger.Errorf("Error creating workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create workflow",
		})
		return
	}
	h.auditLogger.LogWorkflowCreate(
		c.Request.Context(),
		userID,
		workflow.ID,
		c.ClientIP(),
		c.Request.UserAgent(),
		workflow,
	)
	h.logger.Infof("Workflow created: %s by user %s", workflow.ID, userID)
	c.JSON(http.StatusCreated, WorkflowResponse{
		ID:          workflow.ID,
		Name:        workflow.Name,
		Description: workflow.Description,
		Definition:  workflow.Definition,
		Status:      workflow.Status,
		IsActive:    workflow.IsActive,
		CreatedAt:   workflow.CreatedAt,
		UpdatedAt:   workflow.UpdatedAt,
	})
}

// Lấy chi tiết 1 workflow
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	workflowID := c.Param("id")

	workflow, err := h.WorkflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil {
		h.logger.Errorf("Error getting workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get workflow",
		})
		return
	}

	if workflow == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	// Check ownership
	if workflow.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	c.JSON(http.StatusOK, WorkflowResponse{
		ID:          workflow.ID,
		Name:        workflow.Name,
		Description: workflow.Description,
		Definition:  workflow.Definition,
		Status:      workflow.Status,
		IsActive:    workflow.IsActive,
		CreatedAt:   workflow.CreatedAt,
		UpdatedAt:   workflow.UpdatedAt,
	})
}

// Update Workflow
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	workflowID := c.Param("id")
	var req UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request format",
			"details": err.Error(),
		})
		return
	}
	//kiem tra workflow co ton tai
	workflow, err := h.WorkflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil {
		h.logger.Errorf("Error getting workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get workflow",
		})
		return
	}

	if workflow == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}
	//kiem tra dung ID
	if workflow.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}
	//chi update nhung field can thiet
	oldWorkflow := *workflow
	if req.Name != "" {
		workflow.Name = req.Name
	}
	if req.Description != "" {
		workflow.Description = req.Description
	}
	if req.Definition != nil {
		if err := h.ValidateWorkflowDefinition(req.Definition); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid workflow definition",
				"details": err.Error(),
			})
			return
		}
		workflow.Definition = *req.Definition
	}
	if req.IsActive != nil {
		workflow.IsActive = *req.IsActive
		if *req.IsActive {
			workflow.Status = "active"
		} else {
			workflow.Status = "inactive"
		}
	}

	workflow.UpdatedAt = time.Now()
	// Save to database
	if err := h.WorkflowRepo.Update(c.Request.Context(), workflow); err != nil {
		h.logger.Errorf("Error updating workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update workflow",
		})
		return
	}
	h.auditLogger.LogWorkflowUpdate(
		c.Request.Context(),
		userID,
		workflowID,
		c.ClientIP(),
		c.Request.UserAgent(),
		&oldWorkflow,
		workflow,
	)
	h.logger.Infof("Workflow updated: %s", workflow.ID)

	c.JSON(http.StatusOK, WorkflowResponse{
		ID:          workflow.ID,
		Name:        workflow.Name,
		Description: workflow.Description,
		Definition:  workflow.Definition,
		Status:      workflow.Status,
		IsActive:    workflow.IsActive,
		CreatedAt:   workflow.CreatedAt,
		UpdatedAt:   workflow.UpdatedAt,
	})
}

// Xóa workflow
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	workflowID := c.Param("id")

	workflow, err := h.WorkflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil {
		h.logger.Errorf("Error getting workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get workflow",
		})
		return
	}

	if workflow == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	// Check ownership
	if workflow.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	// Delete from databas
	if err := h.WorkflowRepo.Delete(c.Request.Context(), workflowID); err != nil {
		h.logger.Errorf("Error deleting workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete workflow",
		})
		return
	}

	h.logger.Infof("Workflow deleted: %s", workflowID)

	c.JSON(http.StatusOK, gin.H{
		"message": "workflow deleted successfully",
	})
}

// Trigger workflow execution từ webhook hoặc manual
func (h *WorkflowHandler) TriggerWorkflow(c *gin.Context) {
	userID := GetUserID(c)
	workflowID := c.Param("id")

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid payload",
		})
		return
	}

	wf, err := h.WorkflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil || wf == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}
	if wf.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	if !wf.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workflow is not active",
		})
		return
	}

	executionID := uuid.New().String()

	workflowOptions := client.StartWorkflowOptions{
		ID:        executionID, // Unique ID
		TaskQueue: "workflow-queue",
	}

	req := &workflow.ExecuteWorkflowRequest{
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		UserID:      userID,
		TriggerData: payload,
		Definition:  wf.Definition,
	}

	execution, err := h.temporalClient.ExecuteWorkflow(
		c.Request.Context(),
		workflowOptions,
		workflow.ExecuteWorkflowWorkflow, // Workflow function
		req,                              // Input
	)

	if err != nil {
		h.logger.Errorf("Failed to start workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to start workflow",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"execution_id": executionID,
		"workflow_id":  execution.GetID(),
		"run_id":       execution.GetRunID(),
		"status":       "started",
	})
}
func (h *WorkflowHandler) TriggerWorkflowManually(c *gin.Context) {
	userID := GetUserID(c)
	workflowID := c.Param("id")

	h.logger.Infof("Manual trigger for workflow %s by user %s", workflowID, userID)

	wf, err := h.WorkflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil || wf == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	if wf.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	if !wf.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workflow is not active",
		})
		return
	}

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {

		payload = make(map[string]interface{})
	}

	// Add trigger metadata
	payload["_trigger_type"] = "manual"
	payload["_triggered_by"] = userID
	payload["_triggered_at"] = time.Now().Format(time.RFC3339)

	executionID := uuid.New().String()

	workflowOptions := client.StartWorkflowOptions{
		ID:        executionID,
		TaskQueue: "workflow-queue",
	}

	req := &workflow.ExecuteWorkflowRequest{
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		UserID:      userID,
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

	h.logger.Infof("✅ Manual workflow started: %s", executionID)

	c.JSON(http.StatusAccepted, gin.H{
		"success":      true,
		"execution_id": executionID,
		"workflow_id":  execution.GetID(),
		"run_id":       execution.GetRunID(),
		"status":       "processing",
	})
}
func (h *WorkflowHandler) ScheduleWorkflow(c *gin.Context) {
	userID := GetUserID(c)
	workflowID := c.Param("id")

	var req struct {
		Expression string `json:"expression" binding:"required"`
		Timezone   string `json:"timezone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	wf, err := h.WorkflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil || wf == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	if wf.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	// Update workflow definition với cron trigger
	wf.Definition.Trigger = domain.Action{
		ID:   "trigger_cron",
		Type: "cron",
		Config: map[string]interface{}{
			"expression": req.Expression,
			"timezone":   req.Timezone,
		},
	}

	cronConnector := connector.NewCronConnector()
	if err := cronConnector.ValidateConfig(wf.Definition.Trigger.Config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid cron expression: %v", err),
		})
		return
	}

	if err := h.WorkflowRepo.Update(c.Request.Context(), wf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update workflow",
		})
		return
	}

	if err := h.scheduler.ScheduleWorkflow(c.Request.Context(), wf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to schedule: %v", err),
		})
		return
	}

	nextRun := h.scheduler.GetNextRunTime(workflowID)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "workflow scheduled successfully",
		"expression": req.Expression,
		"next_run":   nextRun,
	})
}

func (h *WorkflowHandler) UnscheduleWorkflow(c *gin.Context) {
	userID := GetUserID(c)
	workflowID := c.Param("id")

	wf, err := h.WorkflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil || wf == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	if wf.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	h.scheduler.UnscheduleWorkflow(workflowID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "workflow unscheduled",
	})
}

// WorkflowValidation
func (h *WorkflowHandler) ValidateWorkflowDefinition(def *domain.WorkflowDefinition) error {
	if def.Trigger.Type == "" {
		return fmt.Errorf("Kiểu kích hoạt được yêu cầu")
	}
	if len(def.Actions) == 0 {
		return fmt.Errorf("yêu cầu ít nhất 1 Action ")
	}
	for i, action := range def.Actions {
		if action.Type == "" {
			return fmt.Errorf("Hành động %d: yêu cầu kiểu hành động", i)
		}

	}
	return nil
}

func (h *WorkflowHandler) CreateWorkflowVersion(c *gin.Context) {
	userID := GetUserID(c)
	workflowID := c.Param("id")

	var req struct {
		Definition    domain.WorkflowDefinition `json:"definition" binding:"required"`
		ChangeSummary string                    `json:"change_summary"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	wf, err := h.WorkflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil || wf == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	if wf.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	if err := h.ValidateWorkflowDefinition(&req.Definition); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid definition: %v", err),
		})
		return
	}

	// Create new version
	version := &domain.WorkflowVersion{
		ID:            uuid.New().String(),
		WorkflowID:    workflowID,
		Definition:    req.Definition,
		ChangeSummary: req.ChangeSummary,
		CreatedBy:     userID,
		CreatedAt:     time.Now(),
		IsActive:      false,
	}

	if err := h.versionRepo.Create(c.Request.Context(), version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create version",
		})
		return
	}

	h.logger.Infof("✅ Workflow version created: %s v%d", workflowID, version.Version)

	c.JSON(http.StatusCreated, version)
}
func (h *WorkflowHandler) ListWorkflowVersions(c *gin.Context) {
	userID := GetUserID(c)
	workflowID := c.Param("id")

	wf, err := h.WorkflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil || wf == nil || wf.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	versions, err := h.versionRepo.ListByWorkflowID(c.Request.Context(), workflowID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get versions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"versions": versions,
		"total":    len(versions),
	})
}

// Lấy một version cụ thể
func (h *WorkflowHandler) GetWorkflowVersion(c *gin.Context) {
	userID := GetUserID(c)
	workflowID := c.Param("id")
	versionNum, err := strconv.Atoi(c.Param("version"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid version number",
		})
		return
	}

	wf, err := h.WorkflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil || wf == nil || wf.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	version, err := h.versionRepo.GetByWorkflowAndVersion(
		c.Request.Context(),
		workflowID,
		versionNum,
	)

	if err != nil || version == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "version not found",
		})
		return
	}

	c.JSON(http.StatusOK, version)
}

// Activate một version (rollback/promote)
func (h *WorkflowHandler) ActivateWorkflowVersion(c *gin.Context) {
	userID := GetUserID(c)
	workflowID := c.Param("id")
	versionNum, err := strconv.Atoi(c.Param("version"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid version number",
		})
		return
	}

	wf, err := h.WorkflowRepo.GetByID(c.Request.Context(), workflowID)
	if err != nil || wf == nil || wf.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "workflow not found",
		})
		return
	}

	version, err := h.versionRepo.GetByWorkflowAndVersion(
		c.Request.Context(),
		workflowID,
		versionNum,
	)

	if err != nil || version == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "version not found",
		})
		return
	}

	// Update workflow definition với version này
	wf.Definition = version.Definition
	wf.Version = version.Version
	wf.UpdatedAt = time.Now()

	if err := h.WorkflowRepo.Update(c.Request.Context(), wf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to activate version",
		})
		return
	}

	if err := h.versionRepo.SetActive(c.Request.Context(), workflowID, versionNum); err != nil {
		h.logger.Errorf("Failed to set version active: %v", err)
	}

	h.logger.Infof("✅ Workflow version activated: %s v%d", workflowID, versionNum)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Version %d activated successfully", versionNum),
		"version": versionNum,
	})
}
