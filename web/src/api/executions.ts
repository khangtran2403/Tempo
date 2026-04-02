
import { apiClient } from './client';
import { Execution } from '../types/execution';
import { ExecutionListResponse } from '../types/execution';

export const executionsAPI = {
   // Lấy danh sách executions
  getAll: async (): Promise<Execution[]> => {
    const { data } = await apiClient.get<Execution[]>('/executions');
    return data;
  },
 
  getById: async (id: string): Promise<Execution> => {
    const { data } = await apiClient.get<Execution>(`/executions/${id}`);
    return data;
  },

  listByWorkflowId: async (workflowId: string): Promise<ExecutionListResponse> => {
    const { data } = await apiClient.get<ExecutionListResponse>(`/workflows/${workflowId}/executions`);
    return data;
  },
};