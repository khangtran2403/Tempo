package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/common"
	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

type Logger struct {
	repo   storage.AuditLogRepository
	logger *logrus.Logger
}

func NewLogger(repo storage.AuditLogRepository) *Logger {
	return &Logger{
		repo:   repo,
		logger: common.GetLogger(),
	}
}

// Log ghi audit log
func (l *Logger) Log(ctx context.Context, log *domain.AuditLog) error {
	log.ID = uuid.New().String()
	log.CreatedAt = time.Now()

	if err := l.repo.Create(ctx, log); err != nil {
		l.logger.Errorf("Failed to save audit log: %v", err)
		return err
	}

	l.logger.Infof("📝 Audit: user=%s action=%s resource=%s/%s",
		log.UserID, log.Action, log.Resource, log.ResourceID)

	return nil
}

// LogWorkflowCreate ghi log khi tạo workflow
func (l *Logger) LogWorkflowCreate(
	ctx context.Context,
	userID, workflowID, ipAddress, userAgent string,
	workflow *domain.Workflow,
) error {
	return l.Log(ctx, &domain.AuditLog{
		UserID:     userID,
		Action:     domain.AuditActionCreate,
		Resource:   "workflow",
		ResourceID: workflowID,
		Changes: map[string]interface{}{
			"name":        workflow.Name,
			"description": workflow.Description,
		},
		IPAddress: ipAddress,
		UserAgent: userAgent,
	})
}

// LogWorkflowUpdate ghi log khi update workflow
func (l *Logger) LogWorkflowUpdate(
	ctx context.Context,
	userID, workflowID, ipAddress, userAgent string,
	oldWorkflow, newWorkflow *domain.Workflow,
) error {
	changes := detectChanges(oldWorkflow, newWorkflow)

	return l.Log(ctx, &domain.AuditLog{
		UserID:     userID,
		Action:     domain.AuditActionUpdate,
		Resource:   "workflow",
		ResourceID: workflowID,
		Changes:    changes,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})
}

func (l *Logger) LogWorkflowDelete(
	ctx context.Context,
	userID, workflowID, ipAddress, userAgent string,
) error {
	return l.Log(ctx, &domain.AuditLog{
		UserID:     userID,
		Action:     domain.AuditActionDelete,
		Resource:   "workflow",
		ResourceID: workflowID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})
}

// detectChanges so sánh old và new để tìm changes
func detectChanges(old, new *domain.Workflow) map[string]interface{} {
	changes := make(map[string]interface{})

	if old.Name != new.Name {
		changes["name"] = map[string]string{
			"old": old.Name,
			"new": new.Name,
		}
	}

	if old.Description != new.Description {
		changes["description"] = map[string]string{
			"old": old.Description,
			"new": new.Description,
		}
	}

	if old.IsActive != new.IsActive {
		changes["is_active"] = map[string]bool{
			"old": old.IsActive,
			"new": new.IsActive,
		}
	}

	oldDef, _ := json.Marshal(old.Definition)
	newDef, _ := json.Marshal(new.Definition)
	if string(oldDef) != string(newDef) {
		changes["definition"] = "modified"
	}

	return changes
}
