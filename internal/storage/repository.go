package storage

import (
	"context"

	"github.com/brokeboycoding/tempo/internal/domain"
)

// UserRepository - Interface để thao tác với User
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error

	// GetByID lấy user theo ID
	GetByID(ctx context.Context, id string) (*domain.User, error)

	// GetByEmail lấy user theo email (dùng cho login)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// Update cập nhật user
	Update(ctx context.Context, user *domain.User) error

	// Delete xóa user
	Delete(ctx context.Context, id string) error
}
type WorkflowRepository interface {
	// Create thêm workflow mới
	Create(ctx context.Context, workflow *domain.Workflow) error

	// GetByID lấy workflow theo ID
	GetByID(ctx context.Context, id string) (*domain.Workflow, error)

	// ListByUserID lấy tất cả workflows của user (có pagination)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Workflow, error)

	// Update cập nhật workflow
	Update(ctx context.Context, workflow *domain.Workflow) error

	// Delete xóa workflow
	Delete(ctx context.Context, id string) error

	// GetCount trả về tổng số workflows của user
	GetCount(ctx context.Context, userID string) (int64, error)
}
type ExecutionRepository interface {
	Create(ctx context.Context, exec *domain.WorkflowExecution) error

	GetByID(ctx context.Context, id string) (*domain.WorkflowExecution, error)

	// ListByWorkflowID lấy execution history của một workflow
	ListByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*domain.WorkflowExecution, error)

	Update(ctx context.Context, exec *domain.WorkflowExecution) error

	// GetLatestByWorkflowID lấy execution mới nhất
	GetLatestByWorkflowID(ctx context.Context, workflowID string) (*domain.WorkflowExecution, error)
}

type WebhookHistoryRepository interface {
	Create(ctx context.Context, history *domain.WebhookHistory) error
	GetByID(ctx context.Context, id string) (*domain.WebhookHistory, error)
	ListByWorkflowID(ctx context.Context, workflowID string, limit, offset int) ([]*domain.WebhookHistory, error)
	Delete(ctx context.Context, id string) error
}
type WorkflowVersionRepository interface {
	Create(ctx context.Context, version *domain.WorkflowVersion) error
	GetByID(ctx context.Context, id string) (*domain.WorkflowVersion, error)
	GetByWorkflowAndVersion(ctx context.Context, workflowID string, version int) (*domain.WorkflowVersion, error)
	ListByWorkflowID(ctx context.Context, workflowID string) ([]*domain.WorkflowVersion, error)
	GetLatestVersion(ctx context.Context, workflowID string) (*domain.WorkflowVersion, error)
	SetActive(ctx context.Context, workflowID string, version int) error
}
type AuditLogRepository interface {
	Create(ctx context.Context, audit *domain.AuditLog) error
	ListByUserID(ctx context.Context, userID, resource, resourceID, action string, limit, offset int) ([]*domain.AuditLog, error)
	ListByResource(ctx context.Context, resource, resourceID string) ([]*domain.AuditLog, error)
	LogWorkflowCreate(ctx context.Context, userID, workflowID, ip, userAgent string, workflow interface{}) error
	LogWorkflowUpdate(ctx context.Context, userID, workflowID, ip, userAgent string, oldWorkflow, newWorkflow interface{}) error
}
type IntegrationRepository interface {
	Create(ctx context.Context, integration *domain.Integration) error
	Update(ctx context.Context, integration *domain.Integration) error
	GetByID(ctx context.Context, id string) (*domain.Integration, error)
	ListByUserID(ctx context.Context, userID string) ([]*domain.Integration, error)
	Delete(ctx context.Context, id string) error
}
type SecretRepository interface {
	Create(ctx context.Context, s *domain.Secret) error
	GetByID(ctx context.Context, id string) (*domain.Secret, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Secret, error)
	GetCount(ctx context.Context, userID string) (int64, error)
	Update(ctx context.Context, s *domain.Secret) error
	Delete(ctx context.Context, id string) error
}
