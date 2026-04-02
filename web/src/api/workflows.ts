
import { apiClient } from './client';
import { Workflow } from '../types/workflow';

export const workflowsAPI = {
  
  list: async (page = 1, pageSize = 10) => {
    const { data } = await apiClient.get('/workflows', {
      params: { page, page_size: pageSize },
    });
    return data;
  },

 
  getById: async (id: string): Promise<Workflow> => {
    const { data } = await apiClient.get<Workflow>(`/workflows/${id}`);
    return data;
  },


  create: async (workflow: Partial<Workflow>): Promise<Workflow> => {
    const { data } = await apiClient.post<Workflow>('/workflows', workflow);
    return data;
  },

 
  update: async (id: string, workflow: Partial<Workflow>): Promise<Workflow> => {
    const { data } = await apiClient.put<Workflow>(`/workflows/${id}`, workflow);
    return data;
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/workflows/${id}`);
  },

 
  trigger: async (id: string, payload?: Record<string, any>) => {
    const { data } = await apiClient.post(`/workflows/${id}/trigger`, payload);
    return data;
  },

 
  getExecutions: async (id: string, page = 1, pageSize = 20) => {
    const { data } = await apiClient.get(`/workflows/${id}/executions`, {
      params: { page, page_size: pageSize },
    });
    return data;
  },
};