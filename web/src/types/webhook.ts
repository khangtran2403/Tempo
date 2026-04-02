

export interface WebhookHistory {
  primaryKey: string;
  workflow_id: string;
  method: string;
  headers: Record<string, string>;
  body: Record<string, any>;
  ip_address: string;
  user_agent: string;
  execution_id: string;
  status: 'success' | 'failed';
  error: string;
  received_at: string;
}

