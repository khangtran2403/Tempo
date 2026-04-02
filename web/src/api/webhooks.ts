import { apiClient } from './client';
import { WebhookHistory } from '../types/webhook';

interface WebhookHistoryResponse {
  history: WebhookHistory[];
  page: number;
  page_size: number;
}

interface ReplayResponse {
  success: boolean;
  execution_id: string;
  workflow_id: string;
  message: string;
}

export const webhooksAPI = {
  /**
   * Fetches the webhook call history for a specific workflow.
   * @param workflowId The ID of the workflow.
   * @param page The page number to fetch.
   * @param pageSize The number of items per page.
   */
  listByWorkflowId: async (workflowId: string, page = 1, pageSize = 20): Promise<WebhookHistoryResponse> => {
    const { data } = await apiClient.get<WebhookHistoryResponse>(`/workflows/${workflowId}/webhooks`, {
      params: { page, page_size: pageSize },
    });
    return data;
  },

  /**
   * Triggers a replay of a specific webhook call from history.
   * @param workflowId The ID of the workflow.
   * @param historyId The ID of the webhook history entry to replay.
   */
  replay: async (workflowId: string, historyId: string): Promise<ReplayResponse> => {
    const { data } = await apiClient.post<ReplayResponse>(`/workflows/${workflowId}/webhooks/${historyId}/replay`);
    return data;
  },
};
