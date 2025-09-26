export interface User {
  id: string;
  login: string;
  displayName: string;
  isAdmin: boolean;
  email?: string;
  emailVerifiedAt?: string | null;
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
  email: string;
  emailCode: string;
  password: string;
  confirmPassword: string;
}

export interface LoginRequest {
  login: string;
  password: string;
}

export interface EmailCodeLoginRequest {
  email: string;
  emailCode: string;
}