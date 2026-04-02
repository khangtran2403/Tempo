package domain

import "time"

// WebhookHistory lưu lịch sử webhook calls
type WebhookHistory struct {
	ID          string                 `json:"primaryKey"`
	WorkflowID  string                 `json:"workflow_id"`
	Method      string                 `json:"method"`
	Headers     map[string]string      `json:"headers" gorm:"serializer:json"`
	Body        map[string]interface{} `json:"body" gorm:"serializer:json"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	ExecutionID string                 `json:"execution_id,omitempty"`
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
	ReceivedAt  time.Time              `json:"received_at"`
}
