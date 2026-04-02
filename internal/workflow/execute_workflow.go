package workflow

import (
	"fmt"
	"time"

	"github.com/brokeboycoding/tempo/internal/domain"

	"go.temporal.io/sdk/workflow"
)

// ExecuteWorkflowRequest - Input cho workflow
type ExecuteWorkflowRequest struct {
	WorkflowID  string                    `json:"workflow_id"`  // UUID của workflow definition
	ExecutionID string                    `json:"execution_id"` // UUID cho lần chạy này
	UserID      string                    `json:"user_id"`      // Owner
	TriggerData map[string]interface{}    `json:"trigger_data"` // Data từ webhook/trigger
	Definition  domain.WorkflowDefinition `json:"definition"`   // Workflow config (trigger + actions)
}

// ExecuteWorkflowResponse - Output từ workflow
type ExecuteWorkflowResponse struct {
	ExecutionID string                 `json:"execution_id"`
	Status      string                 `json:"status"` // "success" or "failed"
	Output      map[string]interface{} `json:"output"` // Kết quả từ tất cả actions
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
}

func ExecuteWorkflowWorkflow(ctx workflow.Context, req *ExecuteWorkflowRequest) (*ExecuteWorkflowResponse, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting workflow execution",
		"workflow_id", req.WorkflowID,
		"execution_id", req.ExecutionID)

	startTime := workflow.Now(ctx)
	activityOptions := workflow.ActivityOptions{
		// Timeout: activity phải hoàn thành trong 5 phút
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy:         GetRetryPolicy("linear", 5),
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	response := &ExecuteWorkflowResponse{
		ExecutionID: req.ExecutionID,
		Output:      make(map[string]interface{}),
	}
	logger.Info("Saving execution start to database")
	//luu execution vao DB
	err := workflow.ExecuteActivity(
		ctx,
		"SaveExecutionStart",
		req,
	).Get(ctx, nil)

	if err != nil {
		logger.Error("Failed to save execution start", "error", err)

	}
	//Xac thuc trigger
	logger.Info("Validating trigger", "type", req.Definition.Trigger.Type)

	var triggerOutput interface{}

	err = workflow.ExecuteActivity(
		ctx,
		"ValidateTrigger",
		req.Definition.Trigger,
		req.TriggerData,
	).Get(ctx, &triggerOutput)

	if err != nil {
		logger.Error("Trigger validation failed", "error", err)

		response.Status = "failed"
		response.Error = fmt.Sprintf("trigger failed: %v", err)

		err = workflow.ExecuteActivity(
			ctx,
			"SaveExecutionError",
			req.ExecutionID,
			err.Error(),
		).Get(ctx, nil)

		return response, nil
	}

	response.Output["trigger"] = triggerOutput
	logger.Info("Trigger validated successfully")
	//execute tung action
	logger.Info("Starting actions execution", "count", len(req.Definition.Actions))

	actionResults := make(map[string]interface{})
	actionResults["trigger"] = triggerOutput

	for i, action := range req.Definition.Actions {
		logger.Info("Action info",
			"index", i,
			"action_type", action.Type,
			"action_id", action.ID)
		if action.Condition != nil {
			err = workflow.ExecuteActivity(
				ctx,
				"EvaluateCondition",
				action.Condition,
				actionResults,
			).Get(ctx, nil)

			if err != nil {
				logger.Error("Failed to evaluate condition", "error", err)
				actionResults[action.ID] = map[string]interface{}{
					"skipped": true,
					"reason":  "condition not met",
				}
				continue
			}
		}
		var actionOutput interface{}

		err = workflow.ExecuteActivity(
			ctx,
			"ExecuteActions",
			action,
			actionResults,
		).Get(ctx, &actionOutput)
		if err != nil {
			logger.Error("Action failed",
				"action_id", action.ID,
				"error", err)

			response.Status = "failed"
			response.Error = fmt.Sprintf("action %s failed: %v", action.ID, err)
			response.Output["actions"] = actionResults

			workflow.ExecuteActivity(
				ctx,
				"SaveExecutionError",
				req.ExecutionID,
				err.Error(),
			).Get(ctx, nil)

			return response, nil
		}

		actionResults[action.ID] = actionOutput
		logger.Info("Action completed successfully", "action_id", action.ID)
	}
	response.Output["action"] = actionResults

	response.Status = "success"
	response.Duration = workflow.Now(ctx).Sub(startTime)

	logger.Info("Workflow execution completed",
		"duration", response.Duration,
		"status", response.Status)

	// Save completion to database
	err = workflow.ExecuteActivity(
		ctx,
		"SaveExecutionComplete",
		req.ExecutionID,
		response.Output,
	).Get(ctx, nil)

	if err != nil {
		logger.Error("Failed to save execution complete", "error", err)

	}

	return response, nil
}
