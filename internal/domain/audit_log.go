package domain

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id"`
	Changes    map[string]interface{} `json:"changes" gorm:"serializer:json"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	CreatedAt  time.Time              `json:"created_at"`
}
type WorkflowLog struct {
	ID         string          `json:"id"`
	Action     string          `json:"action"`
	UserID     string          `json:"user_id"`
	WorkflowID string          `json:"workflow_id"`
	IP         string          `json:"ip"`
	UserAgent  string          `json:"user_agent"`
	Metadata   json.RawMessage `json:"metadata"`
	CreatedAt  time.Time       `json:"created_at"`
}
type WorkflowUpdateLog struct {
	ID          string          `json:"id"`
	Action      string          `json:"action"`
	UserID      string          `json:"user_id"`
	WorkflowID  string          `json:"workflow_id"`
	IP          string          `json:"ip"`
	UserAgent   string          `json:"user_agent"`
	OldMetadata json.RawMessage `json:"old_metadata"`
	NewMetadata json.RawMessage `json:"new_metadata"`
	CreatedAt   time.Time       `json:"created_at"`
}

const (
	AuditActionCreate     = "create"
	AuditActionUpdate     = "update"
	AuditActionDelete     = "delete"
	AuditActionActivate   = "activate"
	AuditActionDeactivate = "deactivate"
	AuditActionTrigger    = "trigger"
)
