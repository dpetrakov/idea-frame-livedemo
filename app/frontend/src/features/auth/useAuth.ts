import { useState, useEffect, useCallback } from 'react'
import { authApi } from './api'
import { ApiException } from '../../shared/lib/api'
import type { AuthState, LoginFormData, RegisterFormData, User } from './types'

// Ключи для localStorage
const AUTH_STORAGE_KEY = 'auth_data'

interface StoredAuthData {
  token: string
  user: User
  expiresAt: string
}

// Кастомный хук для управления аутентификацией
export function useAuth() {
  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    token: null,
    expiresAt: null,
    isLoading: true,
    error: null,
  })

  // Сохранение данных в localStorage
  const saveAuthData = useCallback((token: string, user: User, expiresAt: Date) => {
    const data: StoredAuthData = {
      token,
      user,
      expiresAt: expiresAt.toISOString(),
    }
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(data))
  }, [])

  // Очистка данных из localStorage
  const clearAuthData = useCallback(() => {
    localStorage.removeItem(AUTH_STORAGE_KEY)
  }, [])

  // Загрузка данных из localStorage при инициализации
  const loadAuthData = useCallback(() => {
    try {
      const stored = localStorage.getItem(AUTH_STORAGE_KEY)
      if (!stored) {
        setAuthState(prev => ({ ...prev, isLoading: false }))
        return
      }

      const data: StoredAuthData = JSON.parse(stored)
      const expiresAt = new Date(data.expiresAt)

      // Проверяем не истёк ли токен
      if (expiresAt <= new Date()) {
        clearAuthData()
        setAuthState(prev => ({ ...prev, isLoading: false }))
        return
      }

      setAuthState({
        user: data.user,
        token: data.token,
        expiresAt,
        isLoading: false,
        error: null,
      })
    } catch (error) {
      console.error('Error loading auth data:', error)
      clearAuthData()
      setAuthState(prev => ({ ...prev, isLoading: false }))
    }
  }, [clearAuthData])

  // Вход в систему
  const login = useCallback(async (formData: LoginFormData) => {
    setAuthState(prev => ({ ...prev, isLoading: true, error: null }))

    try {
      const response = await authApi.login(formData)
      const expiresAt = new Date(response.expiresAt)

      // Сохраняем данные
      saveAuthData(response.token, response.user, expiresAt)

      setAuthState({
        user: response.user,
        token: response.token,
        expiresAt,
        isLoading: false,
        error: null,
      })

      console.log('Login successful, user authenticated:', response.user)
      return { success: true }
    } catch (error) {
      const message = error instanceof ApiException 
        ? error.data.message 
        : 'Произошла ошибка при входе'

      setAuthState(prev => ({
        ...prev,
        isLoading: false,
        error: message,
      }))

      return { success: false, error: message }
    }
  }, [saveAuthData])

  // Регистрация
  const register = useCallback(async (formData: RegisterFormData) => {
    setAuthState(prev => ({ ...prev, isLoading: true, error: null }))

    try {
      const response = await authApi.register(formData)
      const expiresAt = new Date(response.expiresAt)

      // Сохраняем данные
      saveAuthData(response.token, response.user, expiresAt)

      setAuthState({
        user: response.user,
        token: response.token,
        expiresAt,
        isLoading: false,
        error: null,
      })

      console.log('Registration successful, user authenticated:', response.user)
      return { success: true }
    } catch (error) {
      const message = error instanceof ApiException 
        ? error.data.message 
        : 'Произошла ошибка при регистрации'

      setAuthState(prev => ({
        ...prev,
        isLoading: false,
        error: message,
      }))

      return { success: false, error: message }
    }
  }, [saveAuthData])

  // Выход из системы
  const logout = useCallback(() => {
    clearAuthData()
    setAuthState({
      user: null,
      token: null,
      expiresAt: null,
      isLoading: false,
      error: null,
    })
    console.log('User logged out')
  }, [clearAuthData])

  // Проверка авторизации при загрузке приложения
  useEffect(() => {
    loadAuthData()
  }, [loadAuthData])

  // Автоматический logout при истечении токена
  useEffect(() => {
    if (!authState.expiresAt) return

    const timeUntilExpiry = authState.expiresAt.getTime() - new Date().getTime()
    if (timeUntilExpiry <= 0) {
      logout()
      return
    }

    const timeoutId = setTimeout(() => {
      logout()
    }, timeUntilExpiry)

    return () => clearTimeout(timeoutId)
  }, [authState.expiresAt, logout])

  return {
    ...authState,
    login,
    register,
    logout,
    isAuthenticated: !!authState.user && !!authState.token,
  }
}