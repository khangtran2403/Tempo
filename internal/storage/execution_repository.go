package storage

import (
	"context"
	"errors"

	"github.com/brokeboycoding/tempo/internal/domain"

	"gorm.io/gorm"
)

type PostgresExecutionRepository struct {
	db *gorm.DB
}

func NewPostgresExecutionRepository(db *gorm.DB) ExecutionRepository {
	return &PostgresExecutionRepository{db: db}
}

func (r *PostgresExecutionRepository) Create(ctx context.Context, exec *domain.WorkflowExecution) error {
	return r.db.WithContext(ctx).Create(exec).Error
}

func (r *PostgresExecutionRepository) GetByID(ctx context.Context, id string) (*domain.WorkflowExecution, error) {
	var exec domain.WorkflowExecution
	var ErrNotFound = errors.New("exec not found")
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&exec).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	return &exec, err
}

func (r *PostgresExecutionRepository) ListByWorkflowID(
	ctx context.Context,
	workflowID string,
	limit, offset int,
) ([]*domain.WorkflowExecution, error) {
	var executions []*domain.WorkflowExecution

	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&executions).
		Error

	return executions, err
}

func (r *PostgresExecutionRepository) GetLatestByWorkflowID(
	ctx context.Context,
	workflowID string,
) (*domain.WorkflowExecution, error) {
	var exec domain.WorkflowExecution
	var ErrNotFound = errors.New("exec not found")
	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("started_at DESC").
		First(&exec).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	return &exec, err
}

func (r *PostgresExecutionRepository) Update(ctx context.Context, exec *domain.WorkflowExecution) error {
	return r.db.WithContext(ctx).Save(exec).Error
}
