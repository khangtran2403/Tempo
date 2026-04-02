package storage

import (
	"context"
	"errors"
	"time"

	"github.com/brokeboycoding/tempo/internal/domain"

	"github.com/brokeboycoding/tempo/pkg/metrics"

	"gorm.io/gorm"
)

type PostgresWorkflowRepository struct {
	db *gorm.DB
}

func NewPostgresWorkflowRepository(db *gorm.DB) WorkflowRepository {
	return &PostgresWorkflowRepository{db: db}
}

func (r *PostgresWorkflowRepository) Create(ctx context.Context, workflow *domain.Workflow) error {
	startTime := time.Now()
	err := r.db.WithContext(ctx).Create(workflow).Error
	duration := time.Since(startTime).Seconds()

	metrics.DBQueriesTotal.WithLabelValues(
		"insert",
		"workflows",
	).Inc()

	metrics.DBQueryDuration.WithLabelValues(
		"insert",
		"workflows",
	).Observe(duration)
	return err
}

func (r *PostgresWorkflowRepository) GetByID(ctx context.Context, id string) (*domain.Workflow, error) {
	var workflow domain.Workflow
	var ErrNotFound = errors.New("workflow not found")
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&workflow).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	return &workflow, err
}

// ListByUserID - Danh sách workflows với pagination
func (r *PostgresWorkflowRepository) ListByUserID(
	ctx context.Context,
	userID string,
	limit, offset int,
) ([]*domain.Workflow, error) {
	startTime := time.Now()

	var workflows []*domain.Workflow

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC"). // Mới nhất trước
		Limit(limit).
		Offset(offset).
		Find(&workflows).
		Error
	duration := time.Since(startTime).Seconds()

	metrics.DBQueriesTotal.WithLabelValues(
		"select",
		"workflows",
	).Inc()

	metrics.DBQueryDuration.WithLabelValues(
		"select",
		"workflows",
	).Observe(duration)
	return workflows, err
}

func (r *PostgresWorkflowRepository) GetCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Workflow{}).
		Where("user_id = ?", userID).
		Count(&count).
		Error
	return count, err
}

func (r *PostgresWorkflowRepository) Update(ctx context.Context, workflow *domain.Workflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

func (r *PostgresWorkflowRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&domain.Workflow{}).
		Error
}
