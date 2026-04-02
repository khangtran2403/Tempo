package storage

import (
	"context"
	"errors"

	"github.com/brokeboycoding/tempo/internal/domain"

	"gorm.io/gorm"
)

type IntegrationsRepository struct {
	db *gorm.DB
}

func NewIntegrationsRepository(db *gorm.DB) IntegrationRepository {
	return &IntegrationsRepository{db: db}
}
func (r *IntegrationsRepository) Create(ctx context.Context, integration *domain.Integration) error {
	return r.db.WithContext(ctx).Create(integration).Error
}
func (r *IntegrationsRepository) Update(ctx context.Context, integration *domain.Integration) error {
	// GORM tự động update updated_at
	return r.db.WithContext(ctx).Save(integration).Error
}
func (r *IntegrationsRepository) GetByID(ctx context.Context, id string) (*domain.Integration, error) {
	var integration domain.Integration
	var ErrNotFound = errors.New("integration not found")
	err := r.db.WithContext(ctx).Where("id=?", id).First(&integration).Error
	// GORM trả về gorm.ErrRecordNotFound nếu không tìm thấy
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &integration, nil
}
func (r *IntegrationsRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Integration, error) {
	var integration []*domain.Integration

	q := r.db.WithContext(ctx).Where("user_id=?", userID).Order("created_at DESC")
	if err := q.Find(&integration).Error; err != nil {
		return nil, err
	}
	return integration, nil
}
func (r *IntegrationsRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Integration{}, "id = ?", id).Error
}
