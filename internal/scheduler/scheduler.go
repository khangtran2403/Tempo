package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/internal/workflow"
	"github.com/brokeboycoding/tempo/pkg/common"
)

type Scheduler struct {
	cron           *cron.Cron
	workflowRepo   storage.WorkflowRepository
	temporalClient client.Client
	cfg            *config.Config
	logger         *logrus.Logger
	jobs           map[string]cron.EntryID
	mu             sync.RWMutex
}

func NewScheduler(workflowRepo storage.WorkflowRepository, temporalClient client.Client, cfg *config.Config) *Scheduler {
	return &Scheduler{
		cron:           cron.New(cron.WithParser(cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow))),
		workflowRepo:   workflowRepo,
		temporalClient: temporalClient,
		cfg:            cfg,
		logger:         common.GetLogger(),
		jobs:           make(map[string]cron.EntryID),
		mu:             sync.RWMutex{},
	}
}

// Start khởi động scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.logger.Info("Starting scheduler...")

	if err := s.loadScheduledWorkflows(ctx); err != nil {
		return fmt.Errorf("failed to load workflows: %w", err)
	}
	s.cron.Start()
	s.logger.Infof("✅ Scheduler started with %d jobs", len(s.jobs))

	return nil
}

// Stop dừng scheduler
func (s *Scheduler) Stop() {
	s.logger.Info("Stopping scheduler...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("✅ Scheduler stopped")
}
func (s *Scheduler) loadScheduledWorkflows(ctx context.Context) error {
	// Tạm thời skip, sẽ implement sau
	s.logger.Info("Loading scheduled workflows...")
	return nil
}

// ScheduleWorkflow schedule một workflow
func (s *Scheduler) ScheduleWorkflow(ctx context.Context, wf *domain.Workflow) error {
	if wf.Definition.Trigger.Type != "cron" {
		return fmt.Errorf("workflow trigger is not cron")
	}

	expression, ok := wf.Definition.Trigger.Config["expression"].(string)
	if !ok {
		return fmt.Errorf("missing cron expression")
	}

	s.logger.Infof("Scheduling workflow %s with expression: %s", wf.ID, expression)
	copyWf := &wf
	s.UnscheduleWorkflow(wf.ID)

	entryID, err := s.cron.AddFunc(expression, func() {
		s.triggerWorkflow(context.Background(), *copyWf)
	})

	if err != nil {
		return fmt.Errorf("failed to schedule: %w", err)
	}

	s.mu.Lock()
	s.jobs[wf.ID] = entryID
	s.mu.Unlock()
	s.logger.Infof("✅ Workflow %s scheduled successfully", wf.ID)

	return nil
}

// hủy schedule workflow
func (s *Scheduler) UnscheduleWorkflow(workflowID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entryID, exists := s.jobs[workflowID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, workflowID)
		s.logger.Infof("✅ Workflow %s unscheduled", workflowID)
	}
}
func (s *Scheduler) triggerWorkflow(ctx context.Context, wf *domain.Workflow) {
	s.logger.Infof("⏰ Triggering scheduled workflow: %s", wf.ID)

	executionID := uuid.New().String()

	workflowOptions := client.StartWorkflowOptions{
		ID:        executionID,
		TaskQueue: "workflow-queue",
	}

	req := &workflow.ExecuteWorkflowRequest{
		WorkflowID:  wf.ID,
		ExecutionID: executionID,
		UserID:      wf.UserID,
		TriggerData: map[string]interface{}{
			"trigger_type": "cron",
			"triggered_at": time.Now(),
		},
		Definition: wf.Definition,
	}

	execution, err := s.temporalClient.ExecuteWorkflow(
		ctx,
		workflowOptions,
		workflow.ExecuteWorkflowWorkflow,
		req,
	)

	if err != nil {
		s.logger.Errorf("Failed to trigger workflow %s: %v", wf.ID, err)
		return
	}

	s.logger.Infof("✅ Workflow triggered: execution_id=%s, temporal_id=%s",
		executionID, execution.GetID())
}

// GetScheduledWorkflows trả về danh sách workflows đang được schedule
func (s *Scheduler) GetScheduledWorkflows() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	workflowIDs := make([]string, 0, len(s.jobs))
	for id := range s.jobs {
		workflowIDs = append(workflowIDs, id)
	}
	return workflowIDs
}

// GetNextRunTime lấy thời gian chạy tiếp theo của workflow
func (s *Scheduler) GetNextRunTime(workflowID string) *time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entryID, exists := s.jobs[workflowID]
	if !exists {
		return nil
	}

	entry := s.cron.Entry(entryID)
	if entry.ID == 0 {
		return nil
	}

	nextRun := entry.Next
	return &nextRun
}
