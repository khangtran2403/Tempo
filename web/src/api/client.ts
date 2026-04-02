import axios, { AxiosInstance } from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

class APIClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
      timeout: 15000,
      withCredentials: false,
    });

   
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('token');
        if (token) {
          config.headers = config.headers || {};
          (config.headers as any).Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

   
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        const status = error.response?.status;
        if (status === 401) {
          localStorage.removeItem('token');
          try {
            const current = window.location.pathname;
            if (!current.includes('/login')) {
              window.location.href = '/login';
            }
          } catch {}
        }
        const normalized = {
          message: error.response?.data?.error || error.message || 'Request failed',
          status,
          data: error.response?.data,
        };
        return Promise.reject(normalized);
      }
    );
  }

  getClient() {
    return this.client;
  }
}

export const apiClient = new APIClient().getClient();