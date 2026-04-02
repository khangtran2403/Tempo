package domain

import "time"

type Secret struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	WorkflowID  *string   `json:"workflow_id"`
	Key         string    `json:"key"`
	Value       string    `json:"-"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
