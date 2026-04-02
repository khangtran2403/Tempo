import { apiClient } from "./client";
import { Secret } from "../types/secrets";

export interface CreateSecretRequest {
  Name: string;
  Type: string;
  WorkflowID: string;
  Key: string;
  Value: string;
  Description?: string;
}

export interface UpdateSecretRequest {
  Name?: string;
  Type?: string;
  Key?: string;
  Value?: string;
  Description?: string;
}

export const secretsAPI = {
 
  async create(payload: CreateSecretRequest): Promise<Secret> {
    const res = await apiClient.post("/secrets", payload);
    return res.data;
  },

  async list(workflowId?: string): Promise<Secret[]> {
    const url = workflowId ? `/secrets?workflow_id=${workflowId}` : `/secrets`;
    const res = await apiClient.get(url);
    return Array.isArray(res.data) ? res.data : res.data?.items ?? [];
  },

  
  async get(id: string): Promise<Secret> {
    const res = await apiClient.get(`/secrets/${id}`);
    return res.data;
  },

  
  async update(id: string, payload: UpdateSecretRequest): Promise<Secret> {
    const res = await apiClient.put(`/secrets/${id}`, payload);
    return res.data;
  },

  
  async remove(id: string): Promise<{ success: boolean }> {
    const res = await apiClient.delete(`/secrets/${id}`);
    return res.data;
  }
};

export default secretsAPI;
