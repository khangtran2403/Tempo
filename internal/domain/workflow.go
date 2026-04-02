package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Workflow struct {
	ID          string             `json:"id" gorm:"primaryKey"`
	UserID      string             `json:"user_id" gorm:"index;not null"`
	Name        string             `json:"name" gorm:"not null"`
	Description string             `json:"description"`
	Definition  WorkflowDefinition `json:"definition" gorm:"type:jsonb"`  // JSONB column
	Status      string             `json:"status" gorm:"default:'draft'"` // draft, active, inactive
	IsActive    bool               `json:"is_active" gorm:"default:false"`
	Version     int                `json:"version"`
	ParentID    *string            `json:"parent_id"`
	IsLatest    bool               `json:"is_latest"`
	PublishedAt *time.Time         `json:"published_at"`
	CreatedAt   time.Time          `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt   time.Time          `json:"updated_at" gorm:"autoUpdateTime:milli"`
}

func (Workflow) TableName() string {
	return "workflows"
}

type WorkflowDefinition struct {
	Trigger Action   `json:"trigger"`
	Actions []Action `json:"actions"`
}

// Action là một bước trong workflow
type Action struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Condition *Condition             `json:"condition,omitempty"`
	Transform *Transform             `json:"transform,omitempty"`
}
type Condition struct {
	Field    string      `json:"field"`    // Field từ previous result
	Operator string      `json:"operator"` // ==, !=, >, <, contains
	Value    interface{} `json:"value"`    // Giá trị so sánh
}
type WorkflowVersion struct {
	ID            string             `json:"id"`
	WorkflowID    string             `json:"workflow_id"`
	Version       int                `json:"version"`
	Definition    WorkflowDefinition `json:"definition"`
	ChangeSummary string             `json:"change_summary"`
	CreatedBy     string             `json:"created_by"`
	CreatedAt     time.Time          `json:"created_at"`
	IsActive      bool               `json:"is_active"`
}
type Transform struct {
	// Input mapping từ previous results
	Input map[string]string `json:"input"`
	// Template để transform
	Template string `json:"template"`
	// Output field name
	Output string `json:"output"`
}

// Value() - Khi lưu vào database
func (w WorkflowDefinition) Value() (driver.Value, error) {
	return json.Marshal(w)
}

// Scan() - Khi đọc từ database
func (w *WorkflowDefinition) Scan(value interface{}) error {
	bytes := value.([]byte)
	return json.Unmarshal(bytes, &w)
}
