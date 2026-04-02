import { apiClient } from "./client";
import { WorkflowDefinition } from "../types/workflow";


export interface CreateWorkflowVersionRequest {
  Definition: WorkflowDefinition;
  ChangeSummary?: string;
}

export interface CreateWorkflowVersionResponse {
  ID: string;
  WorkflowID: string;
  Version: string;
  IsActive: boolean;
  CreatedAt: string;
}



export interface WorkflowVersionSummary {
  ID: string;
  WorkflowID: string;
  Version: string;
  CreatedAt: string;
  IsActive: boolean;
}

export interface WorkflowVersionDetail {
  ID: string;
  WorkflowID: string;
  Version: string;
  Definition: WorkflowDefinition;
  ChangeSummary: string;
  CreatedBy: string;
  CreatedAt: string;
  IsActive: boolean;
}



export const workflowVersionAPI = {
 
  async create(
    workflowId: string,
    payload: CreateWorkflowVersionRequest
  ): Promise<CreateWorkflowVersionResponse> {
    const res = await apiClient.post(
      `/workflows/${encodeURIComponent(workflowId)}/versions`,
      payload
    );
    return res.data;
  },

 
  async list(workflowId: string): Promise<WorkflowVersionSummary[]> {
    const res = await apiClient.get(
      `/workflows/${encodeURIComponent(workflowId)}/versions`
    );

    const data = res.data;
    if (Array.isArray(data)) return data as WorkflowVersionSummary[];
    if (Array.isArray(data?.versions))
      return data.versions as WorkflowVersionSummary[];
    return [];
  },

 
  async get(
    workflowId: string,
    version: string | number
  ): Promise<WorkflowVersionDetail> {
    const res = await apiClient.get(
      `/workflows/${encodeURIComponent(workflowId)}/versions/${encodeURIComponent(
        String(version)
      )}`
    );
    return res.data as WorkflowVersionDetail;
  },


  async activate(
    workflowId: string,
    version: string | number
  ): Promise<{ success: boolean; version: string }> {
    const res = await apiClient.post(
      `/workflows/${encodeURIComponent(workflowId)}/versions/${encodeURIComponent(
        String(version)
      )}/activate`
    );
    return res.data;
  },
};

export default workflowVersionAPI;
