package main

import (
	"net/http"
	"strconv"

	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/brokeboycoding/tempo/internal/storage"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditRepo storage.AuditLogRepository
}

func NewAuditHandler(auditRepo storage.AuditLogRepository) *AuditHandler {
	return &AuditHandler{
		auditRepo: auditRepo,
	}
}

// ListAuditLogs - GET /api/v1/audit
// Lấy audit logs (admin only hoặc own logs)
func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	userID := GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	offset := (page - 1) * pageSize

	resource := c.Query("resource")
	resourceID := c.Query("resource_id")
	action := c.Query("action")

	logs, err := h.auditRepo.ListByUserID(
		c.Request.Context(),
		userID,
		resource,
		resourceID,
		action,
		pageSize,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetResourceAuditLogs - GET /api/v1/workflows/:id/audit
// Lấy audit logs cho một resource cụ thể
func (h *AuditHandler) GetResourceAuditLogs(c *gin.Context) {
	userID := GetUserID(c)
	resourceID := c.Param("id")

	logs, err := h.auditRepo.ListByResource(
		c.Request.Context(),
		"workflow",
		resourceID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get audit logs",
		})
		return
	}

	// Filter only user's logs (security)
	filteredLogs := make([]*domain.AuditLog, 0)
	for _, log := range logs {
		if log.UserID == userID {
			filteredLogs = append(filteredLogs, log)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": filteredLogs,
	})
}
