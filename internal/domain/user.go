package domain

import (
	"time"
)

type User struct {
	ID            string      `json:"id" gorm:"primaryKey"`
	Email         string      `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash  string      `json:"-" gorm:"not null"` // json:"-" = không xuất ra JSON
	Name          string      `json:"name"`
	RateLimitTier string      `json:"rate_limit_tier"` // free, pro, enterprise
	CustomLimits  *RateLimits `json:"custom_limits,omitempty" gorm:"type:json"`
	IsActive      bool        `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time   `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt     time.Time   `json:"updated_at" gorm:"autoUpdateTime:milli"`
}
type RateLimits struct {
	RequestsPerMinute  int `json:"requests_per_minute"`
	WorkflowExecutions int `json:"workflow_executions_per_day"`
	MaxActiveWorkflows int `json:"max_active_workflows"`
}

func (u *User) GetRateLimits() *RateLimits {
	if u.CustomLimits != nil {
		return u.CustomLimits
	}

	switch u.RateLimitTier {
	case "enterprise":
		return &RateLimits{
			RequestsPerMinute:  1000,
			WorkflowExecutions: 10000,
			MaxActiveWorkflows: 100,
		}
	case "pro":
		return &RateLimits{
			RequestsPerMinute:  200,
			WorkflowExecutions: 1000,
			MaxActiveWorkflows: 50,
		}
	default: // free
		return &RateLimits{
			RequestsPerMinute:  60,
			WorkflowExecutions: 100,
			MaxActiveWorkflows: 10,
		}
	}
}
func (User) TableName() string {
	return "users"
}
