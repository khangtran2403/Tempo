package main

import (
	"context"
	"os"
	"testing"

	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// This file contains shared mocks and test helpers for the API tests.

// --- Mocks ---

// MockUserRepository mocks storage.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

// MockWorkflowRepository mocks storage.WorkflowRepository
type MockWorkflowRepository struct {
	mock.Mock
}

func (m *MockWorkflowRepository) Create(ctx context.Context, wf *domain.Workflow) error {
	return m.Called(ctx, wf).Error(0)
}
func (m *MockWorkflowRepository) GetByID(ctx context.Context, id string) (*domain.Workflow, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Workflow), args.Error(1)
}
func (m *MockWorkflowRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Workflow, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Workflow), args.Error(1)
}
func (m *MockWorkflowRepository) Update(ctx context.Context, wf *domain.Workflow) error {
	return m.Called(ctx, wf).Error(0)
}
func (m *MockWorkflowRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockWorkflowRepository) GetCount(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// MockWorkflowVersionRepository mocks storage.WorkflowVersionRepository
type MockWorkflowVersionRepository struct {
	mock.Mock
}

func (m *MockWorkflowVersionRepository) Create(ctx context.Context, v *domain.WorkflowVersion) error {
	return m.Called(ctx, v).Error(0)
}
func (m *MockWorkflowVersionRepository) GetByID(ctx context.Context, id string) (*domain.WorkflowVersion, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorkflowVersion), args.Error(1)
}
func (m *MockWorkflowVersionRepository) GetByWorkflowAndVersion(ctx context.Context, workflowID string, version int) (*domain.WorkflowVersion, error) {
	args := m.Called(ctx, workflowID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorkflowVersion), args.Error(1)
}
func (m *MockWorkflowVersionRepository) ListByWorkflowID(ctx context.Context, workflowID string) ([]*domain.WorkflowVersion, error) {
	args := m.Called(ctx, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.WorkflowVersion), args.Error(1)
}
func (m *MockWorkflowVersionRepository) GetLatestVersion(ctx context.Context, workflowID string) (*domain.WorkflowVersion, error) {
	args := m.Called(ctx, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorkflowVersion), args.Error(1)
}
func (m *MockWorkflowVersionRepository) SetActive(ctx context.Context, workflowID string, version int) error {
	return m.Called(ctx, workflowID, version).Error(0)
}

// MockAuditLogRepository mocks storage.AuditLogRepository
type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Create(ctx context.Context, audit *domain.AuditLog) error {
	return m.Called(ctx, audit).Error(0)
}
func (m *MockAuditLogRepository) ListByUserID(ctx context.Context, userID, resource, resourceID, action string, limit, offset int) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, userID, resource, resourceID, action, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}
func (m *MockAuditLogRepository) ListByResource(ctx context.Context, resource, resourceID string) ([]*domain.AuditLog, error) {
	args := m.Called(ctx, resource, resourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AuditLog), args.Error(1)
}
func (m *MockAuditLogRepository) LogWorkflowCreate(ctx context.Context, userID, workflowID, ip, userAgent string, workflow interface{}) error {
	return m.Called(ctx, userID, workflowID, ip, userAgent, workflow).Error(0)
}
func (m *MockAuditLogRepository) LogWorkflowUpdate(ctx context.Context, userID, workflowID, ip, userAgent string, old, new interface{}) error {
	return m.Called(ctx, userID, workflowID, ip, userAgent, old, new).Error(0)
}

// Main test entry point to set Gin mode
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
