// Типы для аутентификации, соответствующие OpenAPI схемам

export interface User {
  id: string
  login: string
  displayName: string
  createdAt: string
}

export interface UserBrief {
  id: string
  login: string
  displayName: string
}

export interface UserRegister {
  login: string
  displayName: string
  password: string
  confirmPassword: string
}

export interface UserLogin {
  login: string
  password: string
}

export interface AuthResponse {
  user: User
  token: string
  expiresAt: string
}

// Локальные типы для состояния аутентификации
export interface AuthState {
  user: User | null
  token: string | null
  expiresAt: Date | null
  isLoading: boolean
  error: string | null
}

export interface LoginFormData {
  login: string
  password: string
}

export interface RegisterFormData {
  login: string
  displayName: string
  password: string
  confirmPassword: string
}