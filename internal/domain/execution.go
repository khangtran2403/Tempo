package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// WorkflowExecution đại diện cho một lần chạy workflow
type WorkflowExecution struct {
	ID                  string        `json:"id" gorm:"primaryKey"`
	WorkflowID          string        `json:"workflow_id" gorm:"index;not null"`
	UserID              string        `json:"user_id" gorm:"index;not null"`
	TemporalExecutionID string        `json:"temporal_execution_id"` // Link to Temporal
	Status              string        `json:"status"`                // running, success, failed, timeout
	StartedAt           time.Time     `json:"started_at" gorm:"autoCreateTime:milli;index"`
	CompletedAt         *time.Time    `json:"completed_at"` // nil nếu còn chạy
	DurationMs          *int64        `json:"duration_ms"`  // Tính toán tự động
	ErrorMessage        string        `json:"error_message"`
	InputData           ExecutionData `json:"input_data" gorm:"type:jsonb"`  // Dữ liệu từ trigger
	OutputData          ExecutionData `json:"output_data" gorm:"type:jsonb"` // Output từ actions
}

func (WorkflowExecution) TableName() string {
	return "workflow_executions"
}

// ExecutionData chứa input/output của execution
type ExecutionData map[string]interface{}

// GORM JSONB Marshaling
func (e ExecutionData) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *ExecutionData) Scan(value interface{}) error {
	bytes := value.([]byte)
	return json.Unmarshal(bytes, &e)
}
