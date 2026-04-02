import { apiClient } from './client';

export interface Integration {
  id: string;
  user_id: string;
  provider: string;
  account_name: string;
  status: string;
  created_at: string;
}

export const integrationsAPI = {
  list: async (): Promise<Integration[]> => {
    const { data } = await apiClient.get('/integrations');
    return data.integrations || [];
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/integrations/${id}`);
  },
};
