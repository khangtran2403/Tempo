package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/connector"
	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/metrics"
	"github.com/sirupsen/logrus"
)

type Dependencies struct {
	ExecutionRepo     storage.ExecutionRepository
	WorkflowRepo      storage.WorkflowRepository
	ConnectorRegistry *connector.Registry
	Config            *config.Config
	Logger            *logrus.Logger
}

type Activity struct {
	deps *Dependencies
}

func NewActivity(deps *Dependencies) *Activity {
	return &Activity{
		deps: deps,
	}
}

// renderTemplates recursively processes a config map, resolving templates.
// It uses a brute-force method to ensure data types are restored after templating.
func renderTemplates(
	config map[string]interface{},
	data map[string]interface{},
	logger *logrus.Logger,
) map[string]interface{} {
	if config == nil {
		return make(map[string]interface{})
	}
	result := make(map[string]interface{})

	for key, value := range config {
		switch v := value.(type) {
		case string:
			if strings.Contains(v, "{{") {
				renderedStr, err := TransformData(v, data)
				if err != nil {
					logger.Warnf("Failed to render template for key '%s': %v. Keeping original value.", key, err)
					result[key] = v
					continue
				}

				trimmedRendered := strings.TrimSpace(renderedStr)
				if (strings.HasPrefix(trimmedRendered, "{") && strings.HasSuffix(trimmedRendered, "}")) ||
					(strings.HasPrefix(trimmedRendered, "[") && strings.HasSuffix(trimmedRendered, "]")) {

					var jsonData interface{}
					if err := json.Unmarshal([]byte(trimmedRendered), &jsonData); err == nil {
						result[key] = jsonData
					} else {
						result[key] = renderedStr
					}
				} else {
					result[key] = renderedStr
				}
			} else {
				result[key] = v
			}

		case map[string]interface{}:
			result[key] = renderTemplates(v, data, logger)
		default:
			result[key] = v
		}
	}
	return result
}

func (a *Activity) SaveExecutionStart(ctx context.Context, req *ExecuteWorkflowRequest) error {
	metrics.WorkflowExecutionsActive.Inc()
	execution := &domain.WorkflowExecution{
		ID:         req.ExecutionID,
		WorkflowID: req.WorkflowID,
		UserID:     req.UserID,
		Status:     "running",
		StartedAt:  time.Now(),
		InputData:  req.TriggerData,
	}

	err := a.deps.ExecutionRepo.Create(ctx, execution)
	if err != nil {
		a.deps.Logger.Errorf("Failed to save execution start: %v", err)
		return fmt.Errorf("failed to create execution record: %w", err)
	}

	a.deps.Logger.Infof("✅ Execution start saved: %s", req.ExecutionID)
	return nil
}

func (a *Activity) ValidateTrigger(ctx context.Context, trigger domain.Action, triggerData map[string]interface{}) (interface{}, error) {
	a.deps.Logger.Infof("Validating trigger type: %s", trigger.Type)
	conn, err := a.deps.ConnectorRegistry.GetConnector(trigger.Type)
	if err != nil {
		return nil, fmt.Errorf("connector '%s' not found: %w", trigger.Type, err)
	}
	if err = conn.ValidateConfig(trigger.Config); err != nil {
		return nil, fmt.Errorf("invalid trigger config: %w", err)
	}

	config := trigger.Config
	if config == nil {
		config = make(map[string]interface{})
	}
	config["trigger_data"] = triggerData
	config["timestamp"] = time.Now()

	output, err := conn.Execute(ctx, config, nil)
	if err != nil {
		return nil, fmt.Errorf("trigger execution failed: %w", err)
	}

	a.deps.Logger.Infof("✅ Trigger validated successfully")
	return output, nil
}

func (a *Activity) ExecuteActions(ctx context.Context, action domain.Action, prevResults map[string]interface{}) (interface{}, error) {
	a.deps.Logger.Infof("Executing action: %s (type: %s)", action.ID, action.Type)

	finalConfig := renderTemplates(action.Config, prevResults, a.deps.Logger)

	conn, err := a.deps.ConnectorRegistry.GetConnector(action.Type)
	if err != nil {
		metrics.ActionExecutionsTotal.WithLabelValues(action.Type, "error").Inc()
		return nil, fmt.Errorf("connector '%s' not found: %w", action.Type, err)
	}

	if err := conn.ValidateConfig(finalConfig); err != nil {
		metrics.ActionExecutionsTotal.WithLabelValues(action.Type, "validation_error").Inc()
		return nil, fmt.Errorf("invalid action config for '%s': %w", action.Type, err)
	}

	startTime := time.Now()
	output, err := conn.Execute(ctx, finalConfig, prevResults)
	duration := time.Since(startTime)

	if err != nil {
		metrics.ActionExecutionsTotal.WithLabelValues(action.Type, "failed").Inc()
		a.deps.Logger.Errorf("❌ Action %s failed after %v: %v", action.ID, duration, err)
		return nil, err
	}

	metrics.ActionExecutionsTotal.WithLabelValues(action.Type, "success").Inc()
	metrics.ActionExecutionDuration.WithLabelValues(action.Type).Observe(duration.Seconds())
	a.deps.Logger.Infof("✅ Action %s completed in %v", action.ID, duration)

	return output, nil
}

func (a *Activity) SaveExecutionComplete(ctx context.Context, executionID string, output map[string]interface{}) error {
	a.deps.Logger.Infof("Saving execution complete: %s", executionID)
	execution, err := a.deps.ExecutionRepo.GetByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to get execution: %w", err)
	}
	if execution == nil {
		return fmt.Errorf("execution %s not found", executionID)
	}

	now := time.Now()
	execution.Status = "success"
	execution.CompletedAt = &now
	execution.OutputData = output

	duration := now.Sub(execution.StartedAt).Milliseconds()
	execution.DurationMs = &duration

	err = a.deps.ExecutionRepo.Update(ctx, execution)
	if err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	a.deps.Logger.Infof("✅ Execution complete saved: %s", executionID)
	metrics.WorkflowExecutionsActive.Dec()
	metrics.WorkflowExecutionsTotal.WithLabelValues(execution.WorkflowID, "success").Inc()

	durationSeconds := float64(duration) / 1000.0
	metrics.WorkflowExecutionDuration.WithLabelValues(execution.WorkflowID).Observe(durationSeconds)

	return nil
}

func (a *Activity) SaveExecutionError(ctx context.Context, executionID string, errorMessage string) error {
	a.deps.Logger.Infof("Saving execution error: %s", executionID)
	execution, err := a.deps.ExecutionRepo.GetByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to get execution: %w", err)
	}
	if execution == nil {
		return fmt.Errorf("execution %s not found", executionID)
	}

	now := time.Now()
	execution.Status = "failed"
	execution.CompletedAt = &now
	execution.ErrorMessage = errorMessage

	duration := now.Sub(execution.StartedAt).Milliseconds()
	execution.DurationMs = &duration

	err = a.deps.ExecutionRepo.Update(ctx, execution)
	if err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	a.deps.Logger.Infof("✅ Execution error saved: %s", executionID)
	metrics.WorkflowExecutionsActive.Dec()
	metrics.WorkflowExecutionsTotal.WithLabelValues(execution.WorkflowID, "failed").Inc()

	return nil
}
