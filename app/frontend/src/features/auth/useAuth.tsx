import { createContext, useContext, useState, useCallback, useEffect, ReactNode } from 'react';
import { apiClient } from '../../shared/lib/api-client';
import { authApi } from './api';
import { User, LoginRequest, RegisterRequest, EmailCodeLoginRequest } from './types';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  login: (data: LoginRequest) => Promise<void>;
  loginByEmailCode: (data: EmailCodeLoginRequest) => Promise<void>;
  register: (data: RegisterRequest) => Promise<void>;
  requestEmailCode: (email: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Проверяем токен при загрузке
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      apiClient.setToken(token);
      authApi
        .getCurrentUser()
        .then(setUser)
        .catch(() => {
          localStorage.removeItem('token');
          apiClient.setToken(null);
        })
        .finally(() => setIsLoading(false));
    } else {
      setIsLoading(false);
    }
  }, []);

  const login = useCallback(async (data: LoginRequest) => {
    const response = await authApi.login(data);
    localStorage.setItem('token', response.token);
    apiClient.setToken(response.token);
    setUser(response.user);
  }, []);

  const loginByEmailCode = useCallback(async (data: EmailCodeLoginRequest) => {
    const response = await authApi.loginByEmailCode(data);
    localStorage.setItem('token', response.token);
    apiClient.setToken(response.token);
    setUser(response.user);
  }, []);

  const register = useCallback(async (data: RegisterRequest) => {
    const response = await authApi.register(data);
    localStorage.setItem('token', response.token);
    apiClient.setToken(response.token);
    setUser(response.user);
  }, []);

  const requestEmailCode = useCallback(async (email: string) => {
    await authApi.requestEmailCode(email);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    apiClient.setToken(null);
    setUser(null);
  }, []);

  const value: AuthContextType = {
    user,
    isLoading,
    isAuthenticated: !!user,
    isAdmin: !!user?.isAdmin,
    login,
    loginByEmailCode,
    register,
    requestEmailCode,
    logout,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}