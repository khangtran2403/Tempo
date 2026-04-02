
import { apiClient } from './client';
import { AuthResponse, User } from '../types/user';

export const authAPI = {
  // Register
  register: async (email: string, password: string, name: string): Promise<AuthResponse> => {
    const { data } = await apiClient.post<AuthResponse>('/auth/register', {
      email,
      password,
      name,
    });
    return data;
  },

  // Login
  login: async (email: string, password: string): Promise<AuthResponse> => {
    const { data } = await apiClient.post<AuthResponse>('/auth/login', {
      email,
      password,
    });
    return data;
  },


  getMe: async (): Promise<User> => {
    const { data } = await apiClient.get<User>('/auth/me');
    return data;
  },
};