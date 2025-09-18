import React, { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../features/auth/AuthContext'
import { Button } from '../shared/ui/Button'
import { Card, CardHeader, CardContent } from '../shared/ui/Card'

export function InitiativesList() {
  const { user, logout, isAuthenticated, isLoading } = useAuth()
  const navigate = useNavigate()

  // Принудительная навигация при логауте
  useEffect(() => {
    if (!isAuthenticated && !isLoading) {
      console.log('InitiativesList: User logged out, navigating to login page')
      navigate('/login', { replace: true })
    }
  }, [isAuthenticated, isLoading, navigate])

  const containerStyles: React.CSSProperties = {
    padding: 'var(--space-4)',
    maxWidth: '800px',
    margin: '0 auto',
  }

  const headerStyles: React.CSSProperties = {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 'var(--space-6)',
  }

  const titleStyles: React.CSSProperties = {
    fontSize: 'var(--fs-2xl)',
    fontWeight: 600,
    color: 'var(--color-text)',
  }

  const welcomeStyles: React.CSSProperties = {
    color: 'var(--color-text-muted)',
    fontSize: 'var(--fs-lg)',
    marginBottom: 'var(--space-4)',
  }

  return (
    <div style={containerStyles}>
      <header style={headerStyles}>
        <h1 style={titleStyles}>Инициативы</h1>
        <Button variant="ghost" onClick={() => {
          console.log('Logout button clicked')
          logout()
          navigate('/login', { replace: true })
        }}>
          Выйти
        </Button>
      </header>

      <Card>
        <CardHeader>
          <p style={welcomeStyles}>
            Добро пожаловать, {user?.displayName}!
          </p>
        </CardHeader>
        <CardContent>
          <p>
            Здесь будет список инициатив. 
            Пока что это заглушка для демонстрации работы аутентификации.
          </p>
          <br />
          <p>
            <strong>Ваши данные:</strong>
          </p>
          <ul>
            <li>ID: {user?.id}</li>
            <li>Логин: {user?.login}</li>
            <li>Отображаемое имя: {user?.displayName}</li>
            <li>Дата регистрации: {user?.createdAt && new Date(user.createdAt).toLocaleDateString('ru-RU')}</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  )
}