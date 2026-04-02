package storage

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/brokeboycoding/tempo/internal/domain"
)

type PostgresSecretRepository struct {
	db *gorm.DB
}

func NewPostgresSecretRepository(db *gorm.DB) SecretRepository {
	return &PostgresSecretRepository{db: db}
}

func (r *PostgresSecretRepository) Create(ctx context.Context, s *domain.Secret) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *PostgresSecretRepository) GetByID(ctx context.Context, id string) (*domain.Secret, error) {
	var s domain.Secret
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresSecretRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Secret, error) {
	var list []*domain.Secret
	q := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Offset(offset)
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PostgresSecretRepository) GetCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Secret{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresSecretRepository) Update(ctx context.Context, s *domain.Secret) error {
	s.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Model(&domain.Secret{}).Where("id = ?", s.ID).Updates(map[string]interface{}{
		"name":       s.Name,
		"type":       s.Type,
		"value":      s.Value,
		"updated_at": s.UpdatedAt,
	}).Error
}

func (r *PostgresSecretRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Secret{}).Error
}
