package workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type RetryStrategy string

const (
	RetryStrategyLinear      RetryStrategy = "linear"
	RetryStrategyExponential RetryStrategy = "exponential"
	RetryStrategyFixed       RetryStrategy = "fixed"
	RetryStrategyNone        RetryStrategy = "none"
)

// GetRetryPolicy trả về retry policy theo strategy
func GetRetryPolicy(strategy RetryStrategy, maxAttempts int) *temporal.RetryPolicy {
	switch strategy {
	case RetryStrategyLinear:
		return &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1.0, // Linear: không tăng
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    int32(maxAttempts),
		}

	case RetryStrategyExponential:
		return &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0, // Exponential: tăng gấp đôi
			MaximumInterval:    time.Minute,
			MaximumAttempts:    int32(maxAttempts),
		}

	case RetryStrategyFixed:
		return &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1.0,
			MaximumInterval:    time.Second, // Giữ nguyên interval
			MaximumAttempts:    int32(maxAttempts),
		}

	case RetryStrategyNone:
		return &temporal.RetryPolicy{
			MaximumAttempts: 1, // Chỉ thử 1 lần
		}

	default:

		return &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		}
	}
}

// WithRetryPolicy apply retry policy cho activity context
func WithRetryPolicy(ctx workflow.Context, strategy RetryStrategy, maxAttempts int) workflow.Context {
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy:         GetRetryPolicy(strategy, maxAttempts),
	}

	return workflow.WithActivityOptions(ctx, activityOptions)
}
