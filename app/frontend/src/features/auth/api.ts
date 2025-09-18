import { apiClient } from '../../shared/lib/api-client';
import { AuthResponse, LoginRequest, RegisterRequest, User } from './types';

export const authApi = {
  async register(data: RegisterRequest): Promise<AuthResponse> {
    return apiClient.post<AuthResponse>('/v1/auth/register', data);
  },

  async login(data: LoginRequest): Promise<AuthResponse> {
    return apiClient.post<AuthResponse>('/v1/auth/login', data);
  },

  async getCurrentUser(): Promise<User> {
    return apiClient.get<User>('/v1/users/me');
  },
};
