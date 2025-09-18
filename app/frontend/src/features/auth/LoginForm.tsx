import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '../../shared/ui/Button'
import { Input } from '../../shared/ui/Input'
import { Card, CardContent, CardFooter, CardHeader } from '../../shared/ui/Card'
import { useAuth } from './AuthContext'
import type { RegisterFormData } from './types'

type FormMode = 'login' | 'register'

export function LoginForm() {
  const [mode, setMode] = useState<FormMode>('login')
  const [formData, setFormData] = useState<RegisterFormData>({
    login: '',
    displayName: '',
    password: '',
    confirmPassword: '',
  })
  const [errors, setErrors] = useState<Partial<Record<keyof RegisterFormData, string>>>({})

  const { login, register, isLoading, error, isAuthenticated } = useAuth()
  const navigate = useNavigate()
  
  // Дебаг-логирование изменений статуса авторизации
  useEffect(() => {
    console.log('LoginForm - isAuthenticated changed:', isAuthenticated, 'isLoading:', isLoading)
  }, [isAuthenticated, isLoading])

  // Принудительная навигация при успешной авторизации
  useEffect(() => {
    if (isAuthenticated && !isLoading) {
      console.log('LoginForm: User authenticated, navigating to home page')
      navigate('/', { replace: true })
    }
  }, [isAuthenticated, isLoading, navigate])

  // Валидация формы
  const validateForm = (): boolean => {
    const newErrors: Partial<Record<keyof RegisterFormData, string>> = {}

    // Валидация логина
    if (!formData.login.trim()) {
      newErrors.login = 'Логин обязателен'
    } else if (formData.login.length < 3 || formData.login.length > 32) {
      newErrors.login = 'Логин должен содержать от 3 до 32 символов'
    } else if (!/^[a-zA-Z0-9_-]+$/.test(formData.login)) {
      newErrors.login = 'Логин может содержать только буквы, цифры, _ и -'
    }

    // Валидация отображаемого имени (только для регистрации)
    if (mode === 'register') {
      if (!formData.displayName.trim()) {
        newErrors.displayName = 'Отображаемое имя обязательно'
      } else if (formData.displayName.length < 1 || formData.displayName.length > 32) {
        newErrors.displayName = 'Отображаемое имя должно содержать от 1 до 32 символов'
      }
    }

    // Валидация пароля
    if (!formData.password) {
      newErrors.password = 'Пароль обязателен'
    } else if (formData.password.length < 8 || formData.password.length > 64) {
      newErrors.password = 'Пароль должен содержать от 8 до 64 символов'
    }

    // Валидация подтверждения пароля (только для регистрации)
    if (mode === 'register') {
      if (!formData.confirmPassword) {
        newErrors.confirmPassword = 'Подтверждение пароля обязательно'
      } else if (formData.password !== formData.confirmPassword) {
        newErrors.confirmPassword = 'Пароли не совпадают'
      }
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  // Обработка отправки формы
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!validateForm()) {
      return
    }

    const result = mode === 'login' 
      ? await login({ login: formData.login, password: formData.password })
      : await register(formData)

    if (result.success) {
      console.log(`${mode === 'login' ? 'Login' : 'Registration'} successful! Navigating immediately...`)
      // Прямая навигация сразу после успешного ответа
      navigate('/', { replace: true })
    }
  }

  // Обработка изменения полей
  const handleChange = (field: keyof RegisterFormData) => (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    setFormData(prev => ({ ...prev, [field]: e.target.value }))
    // Очищаем ошибку для поля при изменении
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }))
    }
  }

  // Переключение режима формы
  const toggleMode = () => {
    setMode(prev => prev === 'login' ? 'register' : 'login')
    setErrors({})
  }

  const containerStyles: React.CSSProperties = {
    display: 'flex',
    minHeight: '100vh',
    alignItems: 'center',
    justifyContent: 'center',
    padding: 'var(--space-4)',
    background: 'var(--color-bg-soft)',
  }

  const formStyles: React.CSSProperties = {
    width: '100%',
    maxWidth: '400px',
  }

  const titleStyles: React.CSSProperties = {
    textAlign: 'center',
    marginBottom: 'var(--space-6)',
    fontSize: 'var(--fs-2xl)',
    fontWeight: 600,
    color: 'var(--color-text)',
  }

  const switchButtonStyles: React.CSSProperties = {
    background: 'none',
    border: 'none',
    color: 'var(--color-highlight)',
    textDecoration: 'underline',
    cursor: 'pointer',
    fontSize: 'var(--fs-sm)',
  }

  const errorMessageStyles: React.CSSProperties = {
    color: 'var(--color-danger)',
    fontSize: 'var(--fs-sm)',
    marginBottom: 'var(--space-3)',
    textAlign: 'center',
  }

  const fieldSpacing: React.CSSProperties = {
    marginBottom: 'var(--space-4)',
  }

  return (
    <div style={containerStyles}>
      <div style={formStyles}>
        <h1 style={titleStyles}>
          {mode === 'login' ? 'Вход в систему' : 'Регистрация'}
        </h1>

        <Card>
          <form onSubmit={handleSubmit}>
            <CardHeader>
              <div style={{ textAlign: 'center' }}>
                {mode === 'login' ? (
                  <p>
                    Нет аккаунта?{' '}
                    <button
                      type="button"
                      onClick={toggleMode}
                      style={switchButtonStyles}
                    >
                      Зарегистрироваться
                    </button>
                  </p>
                ) : (
                  <p>
                    Уже есть аккаунт?{' '}
                    <button
                      type="button"
                      onClick={toggleMode}
                      style={switchButtonStyles}
                    >
                      Войти
                    </button>
                  </p>
                )}
              </div>
            </CardHeader>

            <CardContent>
              {error && (
                <div style={errorMessageStyles}>
                  {error}
                </div>
              )}

              <div style={fieldSpacing}>
                <Input
                  label="Логин"
                  type="text"
                  value={formData.login}
                  onChange={handleChange('login')}
                  error={errors.login}
                  placeholder="Введите логин"
                  autoComplete="username"
                  required
                />
              </div>

              {mode === 'register' && (
                <div style={fieldSpacing}>
                  <Input
                    label="Отображаемое имя"
                    type="text"
                    value={formData.displayName}
                    onChange={handleChange('displayName')}
                    error={errors.displayName}
                    placeholder="Как к вам обращаться?"
                    required
                  />
                </div>
              )}

              <div style={fieldSpacing}>
                <Input
                  label="Пароль"
                  type="password"
                  value={formData.password}
                  onChange={handleChange('password')}
                  error={errors.password}
                  placeholder="Введите пароль"
                  autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
                  required
                />
              </div>

              {mode === 'register' && (
                <div style={fieldSpacing}>
                  <Input
                    label="Подтверждение пароля"
                    type="password"
                    value={formData.confirmPassword}
                    onChange={handleChange('confirmPassword')}
                    error={errors.confirmPassword}
                    placeholder="Повторите пароль"
                    autoComplete="new-password"
                    required
                  />
                </div>
              )}
            </CardContent>

            <CardFooter>
              <Button
                type="submit"
                loading={isLoading}
                style={{ width: '100%' }}
              >
                {mode === 'login' ? 'Войти' : 'Зарегистрироваться'}
              </Button>
            </CardFooter>
          </form>
        </Card>
      </div>
    </div>
  )
}