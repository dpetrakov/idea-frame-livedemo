import React from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { Login } from '../pages/Login'
import { InitiativesList } from '../pages/InitiativesList'
import { useAuth } from '../features/auth/AuthContext'

// Компонент для защищенных маршрутов
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading, user } = useAuth()
  
  console.log('ProtectedRoute - isLoading:', isLoading, 'isAuthenticated:', isAuthenticated, 'user:', user?.login)

  if (isLoading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
        fontSize: 'var(--fs-lg)',
        color: 'var(--color-text-muted)',
      }}>
        Загрузка...
      </div>
    )
  }

  if (!isAuthenticated) {
    console.log('ProtectedRoute: User not authenticated, redirecting to login')
    return <Navigate to="/login" replace />
  }

  console.log('ProtectedRoute: User authenticated, showing protected content')
  return <>{children}</>
}

// Компонент для публичных маршрутов (только для неавторизованных)
function PublicRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading, user } = useAuth()
  
  console.log('PublicRoute - isLoading:', isLoading, 'isAuthenticated:', isAuthenticated, 'user:', user?.login)

  if (isLoading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
        fontSize: 'var(--fs-lg)',
        color: 'var(--color-text-muted)',
      }}>
        Загрузка...
      </div>
    )
  }

  if (isAuthenticated) {
    console.log('PublicRoute: User authenticated, redirecting to main page')
    return <Navigate to="/" replace />
  }

  console.log('PublicRoute: User not authenticated, showing public content')
  return <>{children}</>
}

export function AppRoutes() {
  return (
    <Routes>
      {/* Публичные маршруты */}
      <Route path="/login" element={
        <PublicRoute>
          <Login />
        </PublicRoute>
      } />

      {/* Защищенные маршруты */}
      <Route path="/" element={
        <ProtectedRoute>
          <InitiativesList />
        </ProtectedRoute>
      } />

      {/* Заглушки для будущих маршрутов */}
      <Route path="/item/:id" element={
        <ProtectedRoute>
          <div style={{ padding: 'var(--space-4)' }}>
            Страница инициативы (будет реализована позже)
          </div>
        </ProtectedRoute>
      } />

      {/* Fallback для несуществующих маршрутов */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}