import axios, { AxiosError, AxiosInstance, AxiosRequestConfig } from 'axios';

// API Base URL
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // Get token from storage
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('access_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as AxiosRequestConfig & { _retry?: boolean };

    // Handle 401 - Try to refresh token
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          const response = await axios.post(`${API_BASE_URL}/v1/auth/refresh`, {
            refresh_token: refreshToken,
          });

          const { access_token, refresh_token } = response.data;
          localStorage.setItem('access_token', access_token);
          localStorage.setItem('refresh_token', refresh_token);

          // Retry original request
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${access_token}`;
          }
          return api(originalRequest);
        }
      } catch (refreshError) {
        // Refresh failed - clear tokens and redirect to login
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        window.location.href = '/login';
      }
    }

    return Promise.reject(error);
  }
);

// API Error type
export interface ApiError {
  error: string;
  message: string;
  details?: string;
}

// Generic API response
export interface ApiResponse<T> {
  data: T;
  status: number;
}

// Auth types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface User {
  id: number;
  email: string;
  role: string;
  is_active: boolean;
  email_verified_at?: string;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

// Captcha types
export interface SolveRequest {
  image_base64: string;
  model?: string;
  preprocess?: {
    grayscale?: boolean;
    threshold?: number;
    denoise?: boolean;
  };
}

export interface SolveResponse {
  text: string;
  confidence: number;
  model: string;
  processing_time_ms: number;
}

export interface CaptchaModel {
  id: number;
  name: string;
  type: string;
  version: string;
  accuracy?: number;
  is_active: boolean;
  is_default: boolean;
  description?: string;
  created_at: string;
}

// API Key types
export interface ApiKey {
  id: number;
  name: string;
  key_prefix: string;
  rate_limit: number;
  total_requests: number;
  last_used_at?: string;
  is_active: boolean;
  expires_at?: string;
  created_at: string;
}

export interface CreateApiKeyRequest {
  name: string;
  rate_limit?: number;
  expires_in?: number;
  scopes?: string[];
}

export interface CreateApiKeyResponse extends ApiKey {
  key: string;
}

// Stats types
export interface Stats {
  total_requests: number;
  successful_requests: number;
  failed_requests: number;
  average_processing_time_ms: number;
  accuracy_rate: number;
  models_count: number;
  active_models_count: number;
}

// Auth API
export const authApi = {
  login: (data: LoginRequest) =>
    api.post<AuthResponse>('/v1/auth/login', data),

  register: (data: RegisterRequest) =>
    api.post<User>('/v1/auth/register', data),

  logout: () =>
    api.post('/v1/auth/logout'),

  refresh: (refreshToken: string) =>
    api.post<AuthResponse>('/v1/auth/refresh', { refresh_token: refreshToken }),

  getMe: () =>
    api.get<User>('/v1/auth/me'),

  updateMe: (data: Partial<User>) =>
    api.put<User>('/v1/auth/me', data),

  changePassword: (currentPassword: string, newPassword: string) =>
    api.put('/v1/auth/me/password', {
      current_password: currentPassword,
      new_password: newPassword,
    }),
};

// Captcha API
export const captchaApi = {
  solve: (data: SolveRequest) =>
    api.post<SolveResponse>('/v1/captcha/solve', data),

  solveBatch: (images: SolveRequest[]) =>
    api.post('/v1/captcha/solve/batch', { images }),

  getModels: () =>
    api.get<CaptchaModel[]>('/v1/captcha/models'),

  uploadModel: (formData: FormData) =>
    api.post('/v1/captcha/models/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),

  getStats: () =>
    api.get<Stats>('/v1/captcha/stats'),

  getLogs: (params?: { page?: number; limit?: number; model_id?: number }) =>
    api.get('/v1/captcha/logs', { params }),
};

// API Keys API
export const apiKeysApi = {
  list: () =>
    api.get<ApiKey[]>('/v1/api-keys'),

  create: (data: CreateApiKeyRequest) =>
    api.post<CreateApiKeyResponse>('/v1/api-keys', data),

  delete: (id: number) =>
    api.delete(`/v1/api-keys/${id}`),
};

// Training API
export const trainingApi = {
  start: (data: {
    name: string;
    model_type: string;
    config: {
      epochs: number;
      batch_size: number;
      learning_rate: number;
      validation_split: number;
    };
    dataset_path?: string;
  }) => api.post('/v1/captcha/train', data),

  getStatus: (jobId: number) =>
    api.get(`/v1/captcha/train/${jobId}`),
};

export default api;