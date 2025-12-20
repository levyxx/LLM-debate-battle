import axios from 'axios';
import type {
  LoginRequest,
  LoginResponse,
  User,
  CreateDebateRequest,
  CreateDebateResponse,
  SendMessageRequest,
  SendMessageResponse,
  EndDebateResponse,
  LLMDebateStepResponse,
  DebateHistoryResponse,
  UserStats,
  DebateSession,
  DebateTopicInfo,
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// リクエストインターセプター：認証トークンを追加
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// レスポンスインターセプター：認証エラーの処理
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// 認証API
export const authApi = {
  register: async (username: string, password: string): Promise<User> => {
    const response = await api.post<User>('/api/auth/register', { username, password });
    return response.data;
  },

  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/api/auth/login', data);
    return response.data;
  },

  logout: async (): Promise<void> => {
    await api.post('/api/auth/logout');
  },

  getCurrentUser: async (): Promise<User> => {
    const response = await api.get<User>('/api/auth/me');
    return response.data;
  },
};

// ディベートAPI
export const debateApi = {
  generateTopic: async (): Promise<DebateTopicInfo> => {
    const response = await api.post<DebateTopicInfo>('/api/debate/generate-topic');
    return response.data;
  },

  createDebate: async (data: CreateDebateRequest): Promise<CreateDebateResponse> => {
    const response = await api.post<CreateDebateResponse>('/api/debate/create', data);
    return response.data;
  },

  sendMessage: async (data: SendMessageRequest): Promise<SendMessageResponse> => {
    const response = await api.post<SendMessageResponse>('/api/debate/message', data);
    return response.data;
  },

  endDebate: async (sessionId: number): Promise<EndDebateResponse> => {
    const response = await api.post<EndDebateResponse>('/api/debate/end', { session_id: sessionId });
    return response.data;
  },

  llmDebateStep: async (sessionId: number): Promise<LLMDebateStepResponse> => {
    const response = await api.post<LLMDebateStepResponse>('/api/debate/llm-step', { session_id: sessionId });
    return response.data;
  },

  getDebate: async (id: number): Promise<DebateHistoryResponse> => {
    const response = await api.get<DebateHistoryResponse>(`/api/debate/${id}`);
    return response.data;
  },
};

// ユーザーAPI
export const userApi = {
  getStats: async (): Promise<UserStats> => {
    const response = await api.get<UserStats>('/api/user/stats');
    return response.data;
  },

  getHistory: async (): Promise<DebateSession[]> => {
    const response = await api.get<DebateSession[]>('/api/user/history');
    return response.data;
  },
};

export default api;
