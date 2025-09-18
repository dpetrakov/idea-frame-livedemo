import { apiClient } from '../../shared/lib/api-client';
import { AuthResponse, LoginRequest, RegisterRequest, User } from './types';

export const authApi = {
  async register(data: RegisterRequest): Promise<AuthResponse> {
    return apiClient.post<AuthResponse>('/auth/register', data);
  },

  async login(data: LoginRequest): Promise<AuthResponse> {
    return apiClient.post<AuthResponse>('/auth/login', data);
  },

  async getCurrentUser(): Promise<User> {
    return apiClient.get<User>('/users/me');
  },
};