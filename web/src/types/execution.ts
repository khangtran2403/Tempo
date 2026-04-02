// src/types/execution.ts
export interface Execution {
  id: string;
  workflow_id: string;
  status: 'đang chạy' | 'thành công' | 'thất bại' | 'hết thời gian chờ';
  started_at: string;
  completed_at?: string;
  duration_ms?: number;
  error_message?: string;
  input_data?: Record<string, any>;
  output_data?: Record<string, any>;
  temporal_execution_id?: string; // Add this field based on backend ExecutionResponse
}

export interface ExecutionResponse {
  id: string;
  workflow_id: string;
  status: '' | 'đang chạy' | 'thành công' | 'thất bại' | 'hết thời gian chờ';
  started_at: string;
  completed_at?: string;
  duration_ms?: number;
  error_message?: string;
  input_data?: Record<string, any>;
  output_data?: Record<string, any>;
  temporal_execution_id?: string;
}

export interface ExecutionListResponse {
  executions: ExecutionResponse[];
  total_count: number;
  page: number;
  page_size: number;
}
