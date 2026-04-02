package main

import (

	"bytes"

	"encoding/json"

	"net/http"

	"net/http/httptest"

	"testing"



	"github.com/brokeboycoding/tempo/internal/auth"

	"github.com/brokeboycoding/tempo/internal/config"

	"github.com/brokeboycoding/tempo/internal/domain"

	"github.com/brokeboycoding/tempo/internal/scheduler"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"

	"go.temporal.io/sdk/client"

)



// --- Test Setup ---



const testUserID = "test-user-id"



func setupWorkflowTestRouter() (*gin.Engine, *config.Config, *MockWorkflowRepository, *MockAuditLogRepository, *MockWorkflowVersionRepository, client.Client) {

	router := gin.New()



	mockWorkflowRepo := new(MockWorkflowRepository)

	mockVersionRepo := new(MockWorkflowVersionRepository)

	mockAuditLogger := new(MockAuditLogRepository)

	var mockScheduler *scheduler.Scheduler

	var mockTemporalClient client.Client



	cfg := &config.Config{

		JWT: struct{ Secret string }{Secret: "test-secret-for-api"},

	}



	sched := scheduler.NewScheduler(nil, nil, nil)

	mockScheduler = sched



	workflowHandler := NewWorkflowHandler(mockWorkflowRepo, mockVersionRepo, mockAuditLogger, *mockScheduler, mockTemporalClient, cfg)



	protected := router.Group("/workflows")

	protected.Use(AuthMiddleware(cfg))

	{

		protected.POST("", workflowHandler.CreateWorkflow)

		protected.GET("", workflowHandler.ListWorkflows)

		protected.GET("/:id", workflowHandler.GetWorkflow)

		protected.PUT("/:id", workflowHandler.UpdateWorkflow)

		protected.DELETE("/:id", workflowHandler.DeleteWorkflow)

	}



	return router, cfg, mockWorkflowRepo, mockAuditLogger, mockVersionRepo, mockTemporalClient

}



// --- Tests ---



func TestCreateWorkflow(t *testing.T) {

	router, cfg, mockWorkflowRepo, mockAuditLogger, _, _ := setupWorkflowTestRouter()



		mockWorkflowRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Workflow")).Return(nil)



		mockAuditLogger.On("LogWorkflowCreate", mock.Anything, testUserID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)



	



		def := domain.WorkflowDefinition{



			Trigger: domain.Action{Type: "webhook", ID: "trigger"},

		Actions: []domain.Action{{ID: "action1", Type: "http"}},

	}

	reqBody := CreateWorkflowRequest{

		Name:        "My New Workflow",

		Description: "A test workflow",

		Definition:  def,

	}

	bodyBytes, _ := json.Marshal(reqBody)



	testToken, _ := auth.GenerateToken(testUserID, "test@example.com", cfg.JWT.Secret)

	req, _ := http.NewRequest(http.MethodPost, "/workflows", bytes.NewReader(bodyBytes))

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", "Bearer "+testToken)



	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)



	assert.Equal(t, http.StatusCreated, w.Code)

	var respBody WorkflowResponse

	json.Unmarshal(w.Body.Bytes(), &respBody)

	assert.Equal(t, "My New Workflow", respBody.Name)



		mockWorkflowRepo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*domain.Workflow"))



		mockAuditLogger.AssertCalled(t, "LogWorkflowCreate", mock.Anything, testUserID, mock.Anything, mock.Anything, mock.Anything, mock.Anything)



	}



func TestGetWorkflow(t *testing.T) {

	router, cfg, mockWorkflowRepo, _, _, _ := setupWorkflowTestRouter()

	testToken, _ := auth.GenerateToken(testUserID, "test@example.com", cfg.JWT.Secret)

	workflowID := uuid.New().String()



	t.Run("Successful Get", func(t *testing.T) {

		mockWorkflow := &domain.Workflow{ID: workflowID, UserID: testUserID, Name: "Test Workflow"}

		mockWorkflowRepo.On("GetByID", mock.Anything, workflowID).Return(mockWorkflow, nil).Once()



		req, _ := http.NewRequest(http.MethodGet, "/workflows/"+workflowID, nil)

		req.Header.Set("Authorization", "Bearer "+testToken)

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)



		assert.Equal(t, http.StatusOK, w.Code)

		var respBody WorkflowResponse

		json.Unmarshal(w.Body.Bytes(), &respBody)

		assert.Equal(t, workflowID, respBody.ID)

	})



	t.Run("Forbidden", func(t *testing.T) {

		mockWorkflowForbidden := &domain.Workflow{ID: workflowID, UserID: "another-user"}

		mockWorkflowRepo.On("GetByID", mock.Anything, workflowID).Return(mockWorkflowForbidden, nil).Once()



		req, _ := http.NewRequest(http.MethodGet, "/workflows/"+workflowID, nil)

		req.Header.Set("Authorization", "Bearer "+testToken)

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

	})

}



func TestUpdateWorkflow(t *testing.T) {

	router, cfg, mockWorkflowRepo, mockAuditLogger, _, _ := setupWorkflowTestRouter()

	testToken, _ := auth.GenerateToken(testUserID, "test@example.com", cfg.JWT.Secret)

	workflowID := uuid.New().String()



	mockWorkflow := &domain.Workflow{

		ID:     workflowID,

		UserID: testUserID,

		Name:   "Old Name",

		Definition: domain.WorkflowDefinition{Trigger: domain.Action{Type: "webhook"}, Actions: []domain.Action{}},

		}

		mockWorkflowRepo.On("GetByID", mock.Anything, workflowID).Return(mockWorkflow, nil)

		mockWorkflowRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Workflow")).Return(nil)

		mockAuditLogger.On("LogWorkflowUpdate", mock.Anything, testUserID, workflowID, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		updateReq := UpdateWorkflowRequest{Name: "New Name"}

	bodyBytes, _ := json.Marshal(updateReq)

	req, _ := http.NewRequest(http.MethodPut, "/workflows/"+workflowID, bytes.NewReader(bodyBytes))

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", "Bearer "+testToken)



	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)



	assert.Equal(t, http.StatusOK, w.Code)

	var respBody WorkflowResponse

	json.Unmarshal(w.Body.Bytes(), &respBody)

	assert.Equal(t, "New Name", respBody.Name)



		mockWorkflowRepo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*domain.Workflow"))



		mockAuditLogger.AssertCalled(t, "LogWorkflowUpdate", mock.Anything, testUserID, workflowID, mock.Anything, mock.Anything, mock.Anything, mock.Anything)



	}



func TestDeleteWorkflow(t *testing.T) {

	router, cfg, mockWorkflowRepo, _, _, _ := setupWorkflowTestRouter()

	testToken, _ := auth.GenerateToken(testUserID, "test@example.com", cfg.JWT.Secret)

	workflowID := uuid.New().String()



	mockWorkflow := &domain.Workflow{ID: workflowID, UserID: testUserID}

	mockWorkflowRepo.On("GetByID", mock.Anything, workflowID).Return(mockWorkflow, nil)

	mockWorkflowRepo.On("Delete", mock.Anything, workflowID).Return(nil)



	req, _ := http.NewRequest(http.MethodDelete, "/workflows/"+workflowID, nil)

	req.Header.Set("Authorization", "Bearer "+testToken)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)



	assert.Equal(t, http.StatusOK, w.Code)

	mockWorkflowRepo.AssertCalled(t, "Delete", mock.Anything, workflowID)

}
