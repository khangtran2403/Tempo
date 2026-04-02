package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.temporal.io/sdk/worker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/connector"
	"github.com/brokeboycoding/tempo/internal/scheduler"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/internal/temporal"
	"github.com/brokeboycoding/tempo/internal/workflow"
	"github.com/brokeboycoding/tempo/pkg/common"
)

const (
	TaskQueueName = "workflow-queue"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := common.GetLogger()
	logger.Info("Starting temporal worker")
	//connect database
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{})
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	logger.Info("✅ Connected to database")
	executionRepo := storage.NewPostgresExecutionRepository(db)
	workflowRepo := storage.NewPostgresWorkflowRepository(db)
	integrationRepo := storage.NewIntegrationsRepository(db)
	temporalClient, err := temporal.NewTemporalClient(cfg)

	if err != nil {
		logger.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()
	logger.Info("Temporal client created")

	connectorRegistry := connector.NewRegistry()

	// Register connectors
	connectorRegistry.Register(connector.NewWebhookConnector())
	connectorRegistry.Register(connector.NewHTTPConnector())
	connectorRegistry.Register(connector.NewEmailConnector(cfg))
	connectorRegistry.Register(connector.NewGitHubConnector(integrationRepo, cfg.EncryptionKey))
	connectorRegistry.Register(connector.NewDiscordConnector(integrationRepo, cfg.EncryptionKey))
	connectorRegistry.Register(connector.NewExcelConnector(cfg))
	connectorRegistry.Register(connector.NewGoogleSheetsConnector(integrationRepo, cfg.EncryptionKey, cfg))
	connectorRegistry.Register(connector.NewGCSConnector(integrationRepo, cfg.EncryptionKey, cfg))
	connectorRegistry.Register(connector.NewGoogleDriveConnector(integrationRepo, cfg.EncryptionKey, cfg))
	connectorRegistry.Register(connector.NewNotionConnector(integrationRepo, cfg.EncryptionKey))
	logger.Infof("✅ Registered %d connectors", connectorRegistry.Count())

	deps := &workflow.Dependencies{
		ExecutionRepo:     executionRepo,
		WorkflowRepo:      workflowRepo,
		ConnectorRegistry: connectorRegistry,
		Config:            cfg,
		Logger:            logger,
	}
	w := worker.New(temporalClient, TaskQueueName, worker.Options{
		MaxConcurrentActivityExecutionSize:     10, // Max 10 activities cùng lúc
		MaxConcurrentWorkflowTaskExecutionSize: 5,  // Max 5 workflows cùng lúc
	})
	w.RegisterWorkflow(workflow.ExecuteWorkflowWorkflow)

	activities := workflow.NewActivity(deps)
	w.RegisterActivity(activities.ValidateTrigger)
	w.RegisterActivity(activities.ExecuteActions)
	w.RegisterActivity(activities.SaveExecutionStart)
	w.RegisterActivity(activities.SaveExecutionComplete)
	w.RegisterActivity(activities.SaveExecutionError)
	scheduler := scheduler.NewScheduler(workflowRepo, temporalClient, cfg)
	if err := scheduler.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start scheduler:%v", err)
	}
	//bat dau worker
	err = w.Start()
	if err != nil {
		logger.Fatalf("Unable to start worker: %v", err)
	}

	logger.Info("🚀 Worker started successfully")
	logger.Infof("📋 Task Queue: %s", TaskQueueName)
	logger.Info("⏳ Waiting for workflows...")

	// 11. Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down worker...")
	scheduler.Stop()
	w.Stop()
	logger.Info("Worker stopped")
}
