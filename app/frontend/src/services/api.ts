import { UserRegisterRequest, UserLoginRequest, AuthResponse, User, ErrorResponse } from '../types/auth';

const API_BASE_URL = process.env.REACT_APP_API_URL || '/api';

class ApiError extends Error {
  constructor(public response: ErrorResponse, public status: number) {
    super(response.message);
    this.name = 'ApiError';
  }
}

class ApiService {
  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    
    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    // Добавляем авторизационный заголовок если есть токен
    const token = this.getToken();
    if (token && !endpoint.startsWith('/auth/')) {
      config.headers = {
        ...config.headers,
        'Authorization': `Bearer ${token}`,
      };
    }

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        const errorData: ErrorResponse = await response.json();
        throw new ApiError(errorData, response.status);
      }

      return await response.json();
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      
      // Сетевая ошибка или проблема парсинга JSON
      throw new ApiError({
        code: 'NETWORK_ERROR',
        message: 'Network error or server unavailable',
      }, 0);
    }
  }

  // Методы для работы с токеном
  private getToken(): string | null {
    return localStorage.getItem('auth_token');
  }

  setToken(token: string): void {
    localStorage.setItem('auth_token', token);
  }

  clearToken(): void {
    localStorage.removeItem('auth_token');
  }

  // Регистрация
  async register(data: UserRegisterRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    
    this.setToken(response.token);
    return response;
  }

  // Вход
  async login(data: UserLoginRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    
    this.setToken(response.token);
    return response;
  }

  // Получение текущего пользователя
  async getCurrentUser(): Promise<User> {
    return await this.request<User>('/v1/users/me');
  }

  // Выход
  logout(): void {
    this.clearToken();
  }

  // Проверка health
  async health(): Promise<any> {
    return await this.request('/v1/health', {
      headers: {}, // Убираем авторизацию для health check
    });
  }
}

export const apiService = new ApiService();
export { ApiError };