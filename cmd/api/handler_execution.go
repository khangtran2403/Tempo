package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/common"
)

type ExecutionHandler struct {
	ExecutionRepo storage.ExecutionRepository
	WorkflowRepo  storage.WorkflowRepository
	cfg           *config.Config
	logger        *logrus.Logger
}

func NewExecutionHandler(ExecutionRepo storage.ExecutionRepository, WorkflowRepo storage.WorkflowRepository, cfg *config.Config) *ExecutionHandler {
	return &ExecutionHandler{
		ExecutionRepo: ExecutionRepo,
		WorkflowRepo:  WorkflowRepo,
		cfg:           cfg,
		logger:        common.GetLogger(),
	}
}

type ExecutionResponse struct {
	ID                  string                 `json:"id"`
	WorkflowID          string                 `json:"workflow_id"`
	Status              string                 `json:"status"`
	StartedAt           time.Time              `json:"started_at"`
	CompletedAt         *time.Time             `json:"completed_at"`
	DurationMs          *int64                 `json:"duration_ms"`
	ErrorMessage        string                 `json:"error_message,omitempty"`
	InputData           map[string]interface{} `json:"input_data,omitempty"`
	OutputData          map[string]interface{} `json:"output_data,omitempty"`
	TemporalExecutionID string                 `json:"temporal_execution_id,omitempty"`
}

type ExecutionListResponse struct {
	Executions []*ExecutionResponse `json:"executions"`
	TotalCount int64                `json:"total_count"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
}

// Lấy lịch sử chạy của workflow
func (h *ExecutionHandler) ListExecutions(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	workflowID := c.Param("id")

	// Check workflow ownership
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

	if workflow.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	executions, err := h.ExecutionRepo.ListByWorkflowID(
		c.Request.Context(),
		workflowID,
		pageSize,
		offset,
	)
	if err != nil {
		h.logger.Errorf("Error listing executions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list executions",
		})
		return
	}
	executionResponses := make([]*ExecutionResponse, len(executions))
	for i, exec := range executions {
		executionResponses[i] = &ExecutionResponse{
			ID:                  exec.ID,
			WorkflowID:          exec.WorkflowID,
			Status:              exec.Status,
			StartedAt:           exec.StartedAt,
			CompletedAt:         exec.CompletedAt,
			DurationMs:          exec.DurationMs,
			ErrorMessage:        exec.ErrorMessage,
			InputData:           exec.InputData,
			OutputData:          exec.OutputData,
			TemporalExecutionID: exec.TemporalExecutionID,
		}
	}

	c.JSON(http.StatusOK, ExecutionListResponse{
		Executions: executionResponses,
		TotalCount: int64(len(executions)),
		Page:       page,
		PageSize:   pageSize,
	})
}

// Lấy chi tiết 1 execution
func (h *ExecutionHandler) GetExecution(c *gin.Context) {
	userID := GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	executionID := c.Param("id")

	execution, err := h.ExecutionRepo.GetByID(c.Request.Context(), executionID)
	if err != nil {
		h.logger.Errorf("Error getting execution: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get execution",
		})
		return
	}

	if execution == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "execution not found",
		})
		return
	}

	// Check ownership
	workflow, err := h.WorkflowRepo.GetByID(c.Request.Context(), execution.WorkflowID)
	if err != nil || workflow == nil || workflow.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	c.JSON(http.StatusOK, ExecutionResponse{
		ID:                  execution.ID,
		WorkflowID:          execution.WorkflowID,
		Status:              execution.Status,
		StartedAt:           execution.StartedAt,
		CompletedAt:         execution.CompletedAt,
		DurationMs:          execution.DurationMs,
		ErrorMessage:        execution.ErrorMessage,
		InputData:           execution.InputData,
		OutputData:          execution.OutputData,
		TemporalExecutionID: execution.TemporalExecutionID,
	})
}
