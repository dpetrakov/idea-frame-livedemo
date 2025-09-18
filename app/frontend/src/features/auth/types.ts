export interface User {
  id: string;
  login: string;
  displayName: string;
  createdAt: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  expiresAt: string;
}

export interface RegisterRequest {
  login: string;
  displayName: string;
  password: string;
  confirmPassword: string;
}

export interface LoginRequest {
  login: string;
  password: string;
}