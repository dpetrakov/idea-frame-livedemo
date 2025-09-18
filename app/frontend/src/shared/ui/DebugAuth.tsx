import React from 'react'
import { useAuth } from '../../features/auth/AuthContext'

// Отладочный компонент для мониторинга состояния авторизации
export function DebugAuth() {
  const { user, token, isAuthenticated, isLoading, error } = useAuth()

  const debugStyles: React.CSSProperties = {
    position: 'fixed',
    top: '10px',
    right: '10px',
    background: 'rgba(0, 0, 0, 0.8)',
    color: 'white',
    padding: '10px',
    fontSize: '12px',
    borderRadius: '4px',
    zIndex: 9999,
    fontFamily: 'monospace',
    maxWidth: '300px',
    wordBreak: 'break-all',
  }

  return (
    <div style={debugStyles}>
      <div><strong>🔍 Debug Auth State</strong></div>
      <div>isAuthenticated: {isAuthenticated ? '✅' : '❌'}</div>
      <div>isLoading: {isLoading ? '⏳' : '✅'}</div>
      <div>user: {user ? user.login : 'null'}</div>
      <div>token: {token ? `${token.slice(0, 20)}...` : 'null'}</div>
      <div>error: {error || 'null'}</div>
      <div>localStorage: {localStorage.getItem('auth_data') ? '✅' : '❌'}</div>
    </div>
  )
}