package storage

import (
	"context"
	"errors"

	"github.com/brokeboycoding/tempo/internal/domain"

	"gorm.io/gorm"
)

type PostgresUserRepository struct {
	db *gorm.DB
}

func NewPostgreUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	var ErrNotFound = errors.New("user not found")
	err := r.db.WithContext(ctx).Where("id=?", id).First(&user).Error
	// GORM trả về gorm.ErrRecordNotFound nếu không tìm thấy
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil

}

// Lay user by email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	var ErrNotFound = errors.New("user not found")

	err := r.db.WithContext(ctx).Where("email=?", email).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update cập nhật user
func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	// GORM tự động update updated_at
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete user
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id=?", id).Delete(&domain.User{}).Error
}
