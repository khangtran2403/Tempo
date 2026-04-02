
import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'react-hot-toast';


import LoginPage from './pages/auth/LoginPage';
import RegisterPage from './pages/auth/RegisterPage';
import GoogleAuthCallbackPage from './pages/auth/GoogleAuthCallbackPage'; // New import
import DashboardPage from './pages/dashboard/DashboardPage';
import WorkflowsPage from './pages/workflows/WorkflowsPage';
import WorkflowEditorPage from './pages/workflows/WorkflowEditorPage';
import WorkflowDetailPage from './pages/workflows/WorkflowDetailPage';
import CreateWorkflowVersion from './pages/workflowversion/CreateworkflowversionPage';
import WorkflowVersionComparePage from './pages/workflowversion/workflowversionComparePage';
import  WorkflowVersionDetailPage from './pages/workflowversion/workflowversionDetailPage';
import  WorkflowVersionPage from './pages/workflowversion/workflowversionPage';
import CreateSecretPage   from './pages/secrets/CreateSecretPage';
import SecretListPage   from './pages/secrets/SecretListPage';
import UpdateSecretPage   from './pages/secrets/UpdateSecretPage';
import ViewSecretPage   from './pages/secrets/ViewSecretDetailPage';
import SettingPage  from './pages/settings/SettingsPage';
import ExecutionsListPage  from './pages/executions/ExecutionsListPage';
import ExecutionDetailPage  from './pages/executions/ExecutionDetailPage';
import IntegrationsListPage from './pages/integrations/IntegrationsListPage';
import AddIntegrationPage from './pages/integrations/AddIntegrationPage';
import DocumentationPage from './pages/docs/DocumentationPage';
// Components
import ProtectedRoute from './component/common/ProtectedRoutes';
import Layout from './component/layout/Layout';

// Create QueryClient
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          {/* Public routes */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/auth/google/callback" element={<GoogleAuthCallbackPage />} /> {/* New route */}

          {/* Protected routes */}
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<DashboardPage />} />
            <Route path="workflows" element={<WorkflowsPage />} />
            <Route path="workflows/new" element={<WorkflowEditorPage />} />
            <Route path="workflows/:id" element={<WorkflowDetailPage />} />
            <Route path="workflows/:id/edit" element={<WorkflowEditorPage />} />
            <Route path="workflows/:id/versions" element={<WorkflowVersionPage />} />
            <Route path="workflows/:id/versions/new" element={<CreateWorkflowVersion />} />
            <Route path="workflows/:id/versions/compare" element={<WorkflowVersionComparePage />} />
            <Route path="workflows/:id/versions/:version" element={<WorkflowVersionDetailPage />} />
            <Route path="secrets" element={<SecretListPage />} />
            <Route path="secrets/new" element={<CreateSecretPage />} />
            <Route path="secrets/:id" element={<ViewSecretPage />} />
            <Route path="secrets/:id/edit" element={<UpdateSecretPage />} />
            <Route path="settings" element={<SettingPage />} />
            <Route path="integrations" element={<IntegrationsListPage />} />
            <Route path="integrations/add" element={<AddIntegrationPage />} />
            <Route path="documentation" element={<DocumentationPage />} />
            <Route path="executions">
              <Route index element={<ExecutionsListPage />} />
              <Route path=":id" element={<ExecutionDetailPage />} />
            </Route>
          </Route>

          {/* 404 */}
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>

        {/* Toast notifications */}
        <Toaster position="top-right" />
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;