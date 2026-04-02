package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/brokeboycoding/tempo/internal/config"
	"github.com/brokeboycoding/tempo/internal/domain"
	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/common"

	"github.com/brokeboycoding/tempo/internal/scheduler"

	"github.com/brokeboycoding/tempo/internal/temporal"
	"github.com/brokeboycoding/tempo/pkg/ratelimit"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger := common.GetLogger()
	logger.Infof("Starting Tempo API in %s mode", cfg.Environment)

	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{})
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	logger.Info("✅ Connected to database")

	err = db.AutoMigrate(
		&domain.User{}, &domain.Workflow{}, &domain.WorkflowVersion{},
		&domain.WorkflowExecution{}, &domain.WebhookHistory{}, &domain.Secret{},
		&domain.AuditLog{}, &domain.WorkflowLog{}, &domain.WorkflowUpdateLog{},
		&domain.Integration{},
	)
	if err != nil {
		logger.Fatalf("❌ Migration failed: %v", err)
	}
	logger.Info("🚀 Database migrated successfully")

	temporalClient, err := temporal.NewTemporalClient(cfg)
	if err != nil {
		logger.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()
	logger.Info("✅ Temporal client initialized")

	ratelimiter, err := ratelimit.NewRateLimiter()
	if err != nil {
		logger.Fatalf("Failed to create rate limiter: %v", err)
	}
	defer ratelimiter.Close()
	logger.Info("✅ Rate limiter initialized")

	userRepo := storage.NewPostgreUserRepository(db)
	workflowRepo := storage.NewPostgresWorkflowRepository(db)
	versionRepo := storage.NewPostgresWorkflowVersionRepository(db)
	webhookHistoryRepo := storage.NewPostgresWebhookHistoryRepository(db)
	executionRepo := storage.NewPostgresExecutionRepository(db)
	auditLogger := storage.NewPostgresAuditLogRepository(db)
	secretRepo := storage.NewPostgresSecretRepository(db)

	scheduler := scheduler.NewScheduler(workflowRepo, temporalClient, cfg)

	authHandler := NewAuthHandler(userRepo, cfg)
	workflowHandler := NewWorkflowHandler(workflowRepo, versionRepo, auditLogger, *scheduler, temporalClient, cfg)
	executionHandler := NewExecutionHandler(executionRepo, workflowRepo, cfg)
	webhookHandler := NewWebhookHandler(workflowRepo, webhookHistoryRepo, temporalClient, cfg)
	auditHandler := NewAuditHandler(auditLogger)
	secretHandler := NewSecretHandler(secretRepo, cfg)
	integrationRepo := storage.NewIntegrationsRepository(db)
	integrationHandler := NewIntegrationHandler(integrationRepo, cfg)

	// Setup Gin Engine
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(PrometheusMiddleware())
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now()})
	})

	// Setup Routes
	setupRoutes(r, cfg, authHandler, workflowHandler, executionHandler, webhookHandler, auditHandler, secretHandler, integrationHandler)

	// Start HTTP Server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.API.Port),
		Handler: r,
	}

	go func() {
		logger.Infof("API server listening on port %d", cfg.API.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Listen error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

func setupRoutes(r *gin.Engine, cfg *config.Config, authHandler *AuthHandler, workflowHandler *WorkflowHandler, executionHandler *ExecutionHandler, webhookHandler *WebhookHandler, auditHandler *AuditHandler, secretHandler *SecretHandler, integrationHandler *IntegrationHandler) {

	// Public auth routes for user login
	authRoutes := r.Group("/auth")
	{
		authRoutes.GET("/google/login", authHandler.HandleGoogleLogin)
		authRoutes.GET("/google/callback", authHandler.HandleGoogleCallback)
	}

	integrationAuthRoutes := r.Group("/integrations")
	{
		integrationAuthRoutes.GET("/google/connect", AuthMiddleware(cfg), integrationHandler.HandleGoogleConnect)
		integrationAuthRoutes.GET("/github/connect", AuthMiddleware(cfg), integrationHandler.HandleGitHubConnect)
		integrationAuthRoutes.GET("/notion/connect", AuthMiddleware(cfg), integrationHandler.HandleNotionConnect)

		integrationAuthRoutes.GET("/google/callback", integrationHandler.HandleGoogleCallback)
		integrationAuthRoutes.GET("/github/callback", integrationHandler.HandleGitHubCallback)
		integrationAuthRoutes.GET("/notion/callback", integrationHandler.HandleNotionCallback)
	}

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}
		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("/:workflow_id", webhookHandler.TriggerWebhook)
		}
		audit := v1.Group("/audit")
		{
			audit.GET("/:id", auditHandler.ListAuditLogs)
		}

		protected := v1.Group("")
		protected.Use(AuthMiddleware(cfg))
		{
			// Auth
			protected.GET("/auth/me", authHandler.GetMe)

			// Integrations (for listing, deleting - standard API resource)
			integrations := protected.Group("/integrations")
			{
				integrations.GET("", integrationHandler.ListIntegrations)
				integrations.DELETE("/:id", integrationHandler.DeleteIntegration)
			}

			// Workflows
			workflows := protected.Group("/workflows")
			{
				workflows.GET("", workflowHandler.ListWorkflows)
				workflows.POST("", workflowHandler.CreateWorkflow)
				workflows.GET("/:id", workflowHandler.GetWorkflow)
				workflows.PUT("/:id", workflowHandler.UpdateWorkflow)
				workflows.DELETE("/:id", workflowHandler.DeleteWorkflow)
				workflows.POST("/:id/versions", workflowHandler.CreateWorkflowVersion)
				workflows.GET("/:id/versions", workflowHandler.ListWorkflowVersions)
				workflows.GET("/:id/versions/:version", workflowHandler.GetWorkflowVersion)
				workflows.POST("/:id/versions/:version/activate", workflowHandler.ActivateWorkflowVersion)
				workflows.POST("/:id/trigger", workflowHandler.TriggerWorkflowManually)
				workflows.GET("/:id/webhooks", webhookHandler.ListWebhookHistory)
				workflows.POST("/:id/webhooks/:history_id/replay", webhookHandler.ReplayWebhook)
				workflows.POST("/:id/audit", auditHandler.GetResourceAuditLogs)
				workflows.GET("/:id/executions", executionHandler.ListExecutions)
			}

			// Executions
			executions := protected.Group("/executions")
			{
				executions.GET("/:id", executionHandler.GetExecution)
			}
			//Secrets
			secrets := protected.Group("/secrets")
			{
				secrets.POST("", secretHandler.CreateSecret)
				secrets.GET("", secretHandler.ListSecrets)
				secrets.GET("/:id", secretHandler.GetSecret)
				secrets.PUT("/:id", secretHandler.UpdateSecret)
				secrets.DELETE("/:id", secretHandler.DeleteSecret)
			}
		}
	}
}
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
