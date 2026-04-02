package storage

import (
	"context"
	"errors"

	"github.com/brokeboycoding/tempo/internal/domain"

	"gorm.io/gorm"
)

type PostgresWebhookHistoryRepository struct {
	db *gorm.DB
}

func NewPostgresWebhookHistoryRepository(db *gorm.DB) WebhookHistoryRepository {
	return &PostgresWebhookHistoryRepository{db: db}
}

func (r *PostgresWebhookHistoryRepository) Create(
	ctx context.Context,
	history *domain.WebhookHistory,
) error {
	return r.db.WithContext(ctx).Create(history).Error
}
func (r *PostgresWebhookHistoryRepository) GetByID(ctx context.Context, id string) (*domain.WebhookHistory, error) {
	var his domain.WebhookHistory
	var ErrNotFound = errors.New("history not found")
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&his).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	return &his, err
}

func (r *PostgresWebhookHistoryRepository) ListByWorkflowID(
	ctx context.Context,
	workflowID string,
	limit, offset int,
) ([]*domain.WebhookHistory, error) {
	var history []*domain.WebhookHistory

	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("received_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).
		Error

	return history, err
}
func (r *PostgresWebhookHistoryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&domain.WebhookHistory{}).
		Error
}
