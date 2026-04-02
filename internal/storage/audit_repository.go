package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

type PostgresAuditLogRepository struct {
	db *gorm.DB
}

func NewPostgresAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &PostgresAuditLogRepository{db: db}
}

func (r *PostgresAuditLogRepository) Create(
	ctx context.Context,
	audit *domain.AuditLog,
) error {
	return r.db.WithContext(ctx).Create(audit).Error
}

func (r *PostgresAuditLogRepository) ListByUserID(ctx context.Context, userID string, resource string, resourceID string,
	action string,
	pageSize int,
	offset int) ([]*domain.AuditLog, error) {

	var audit []*domain.AuditLog

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC"). // Mới nhất trước
		Limit(pageSize).
		Offset(offset).
		Find(&audit, &resource, &action).
		Error

	return audit, err
}
func (r *PostgresAuditLogRepository) ListByResource(ctx context.Context, string, resourceID string) ([]*domain.AuditLog, error) {
	var audit []*domain.AuditLog

	err := r.db.WithContext(ctx).
		Where("user_id = ?", resourceID).
		Find(&audit).
		Error

	return audit, err
}
func (r *PostgresAuditLogRepository) LogWorkflowCreate(
	ctx context.Context,
	userID string,
	workflowID string,
	ip string,
	userAgent string,
	workflow interface{},
) error {

	meta, err := json.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow metadata: %w", err)
	}

	log := domain.WorkflowLog{
		Action:     "workflow_created",
		UserID:     userID,
		WorkflowID: workflowID,
		IP:         ip,
		UserAgent:  userAgent,
		Metadata:   meta,
		CreatedAt:  time.Now(),
	}

	return r.db.WithContext(ctx).Create(&log).Error
}
func (r *PostgresAuditLogRepository) LogWorkflowUpdate(
	ctx context.Context,
	userID string,
	workflowID string,
	ip string,
	userAgent string,
	oldWorkflow interface{},
	newWorkflow interface{},
) error {

	// JSON hóa dữ liệu old/new
	oldMeta, err := json.Marshal(oldWorkflow)
	if err != nil {
		return fmt.Errorf("failed to marshal old workflow metadata: %w", err)
	}

	newMeta, err := json.Marshal(newWorkflow)
	if err != nil {
		return fmt.Errorf("failed to marshal new workflow metadata: %w", err)
	}

	log := domain.WorkflowUpdateLog{
		ID:          uuid.New().String(),
		Action:      "workflow_updated",
		UserID:      userID,
		WorkflowID:  workflowID,
		IP:          ip,
		UserAgent:   userAgent,
		OldMetadata: oldMeta,
		NewMetadata: newMeta,
		CreatedAt:   time.Now(),
	}

	return r.db.WithContext(ctx).Save(&log).Error
}
