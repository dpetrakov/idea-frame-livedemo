import React from 'react'

interface ErrorBoundaryState {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  ErrorBoundaryState
> {
  constructor(props: { children: React.ReactNode }) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('React Error Boundary caught an error:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100vh',
          padding: 'var(--space-4)',
          textAlign: 'center',
        }}>
          <h1 style={{
            fontSize: 'var(--fs-2xl)',
            color: 'var(--color-danger)',
            marginBottom: 'var(--space-4)',
          }}>
            Произошла ошибка
          </h1>
          <p style={{
            color: 'var(--color-text-muted)',
            marginBottom: 'var(--space-4)',
          }}>
            Приложение столкнулось с неожиданной ошибкой.
          </p>
          <button
            onClick={() => window.location.reload()}
            style={{
              padding: '12px 20px',
              backgroundColor: 'var(--color-highlight)',
              color: 'white',
              border: 'none',
              borderRadius: 'var(--radius)',
              cursor: 'pointer',
              fontSize: 'var(--fs-md)',
            }}
          >
            Перезагрузить страницу
          </button>
        </div>
      )
    }

    return this.props.children
  }
}