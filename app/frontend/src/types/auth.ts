export interface User {
  id: string;
  login: string;
  displayName: string;
  createdAt: string;
}

export interface UserRegisterRequest {
  login: string;
  displayName: string;
  password: string;
  confirmPassword: string;
}

export interface UserLoginRequest {
  login: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  expiresAt: string;
}

export interface ErrorResponse {
  code: string;
  message: string;
  details?: any;
  correlationId?: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}