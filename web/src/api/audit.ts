import { apiClient } from './client';

type AuditLog = {
  id: string;
  user_id: string;
  resource_id: string;
  action: string;
  ip: string;
  user_agent: string;
  created_at: string;
  metadata?: Record<string, any>;
};

type AuditLogListResponse = {
    logs: AuditLog[];
    total_count: number;
    page: number;
    page_size: number;
}

export const auditAPI = {
  getAll: async (): Promise<AuditLogListResponse> => {
    const { data } = await apiClient.get<AuditLogListResponse>('/audit');
    return data;
  },
};
