package storage

import (
	"context"
	"errors"

	"github.com/brokeboycoding/tempo/internal/domain"

	"gorm.io/gorm"
)

type PostgresWorkflowVersionRepository struct {
	db *gorm.DB
}

func NewPostgresWorkflowVersionRepository(db *gorm.DB) WorkflowVersionRepository {
	return &PostgresWorkflowVersionRepository{db: db}
}

func (r *PostgresWorkflowVersionRepository) Create(
	ctx context.Context,
	version *domain.WorkflowVersion,
) error {
	return r.db.WithContext(ctx).Create(version).Error
}

func (r *PostgresWorkflowVersionRepository) ListByWorkflowID(
	ctx context.Context,
	workflowID string,
) ([]*domain.WorkflowVersion, error) {
	var versions []*domain.WorkflowVersion

	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("version DESC").
		Find(&versions).
		Error

	return versions, err
}
func (r *PostgresWorkflowVersionRepository) GetByID(ctx context.Context, id string) (*domain.WorkflowVersion, error) {
	var version domain.WorkflowVersion
	var ErrNotFound = errors.New("workflow version not found")
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&version).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	return &version, err
}
func (r *PostgresWorkflowVersionRepository) GetByWorkflowAndVersion(ctx context.Context, workflowID string, version int) (*domain.WorkflowVersion, error) {
	var wlandver domain.WorkflowVersion

	err := r.db.WithContext(ctx).
		Where("workflow_id = ? AND version = ?", workflowID, version).
		First(&wlandver).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &wlandver, nil
}
func (r *PostgresWorkflowVersionRepository) GetLatestVersion(
	ctx context.Context,
	workflowID string,
) (*domain.WorkflowVersion, error) {
	var version domain.WorkflowVersion

	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("version DESC").
		First(&version).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &version, nil
}

func (r *PostgresWorkflowVersionRepository) SetActive(
	ctx context.Context,
	workflowID string,
	version int,
) error {
	err := r.db.WithContext(ctx).
		Model(&domain.WorkflowVersion{}).
		Where("workflow_id = ?", workflowID).
		Update("is_active", false).
		Error

	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Model(&domain.WorkflowVersion{}).
		Where("workflow_id = ? AND version = ?", workflowID, version).
		Update("is_active", true).
		Error
}
