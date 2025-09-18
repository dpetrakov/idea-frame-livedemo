import { apiClient } from '../../shared/lib/api'
import type { UserLogin, UserRegister, AuthResponse, User } from './types'

// API методы для аутентификации
export const authApi = {
  // Регистрация нового пользователя
  register: async (data: UserRegister): Promise<AuthResponse> => {
    return apiClient.post<AuthResponse>('/auth/register', data)
  },

  // Вход в систему
  login: async (data: UserLogin): Promise<AuthResponse> => {
    return apiClient.post<AuthResponse>('/auth/login', data)
  },

  // Получение информации о текущем пользователе
  getCurrentUser: async (token: string): Promise<User> => {
    return apiClient.get<User>('/users/me', token)
  },

  // Проверка health endpoint
  health: async () => {
    return apiClient.get<{ status: string; database: string; timestamp: string }>('/health')
  },
}