import React from 'react'

export type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
}

export function Button({
  variant = 'primary',
  size = 'md',
  loading = false,
  disabled,
  children,
  style,
  ...props
}: ButtonProps) {
  const baseStyles: React.CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
    borderRadius: 'var(--radius)',
    border: '1px solid transparent',
    boxShadow: 'var(--shadow-sm)',
    fontWeight: 600,
    fontSize: 'var(--fs-sm)',
    cursor: 'pointer',
    transition: 'all 0.2s ease',
    textDecoration: 'none',
    minHeight: '44px', // Минимум для touch targets
  }

  const sizeStyles = {
    sm: { padding: '8px 12px', fontSize: 'var(--fs-xs)' },
    md: { padding: '10px 16px', fontSize: 'var(--fs-sm)' },
    lg: { padding: '12px 20px', fontSize: 'var(--fs-md)' },
  }

  const variantStyles = {
    primary: {
      background: 'var(--color-highlight)',
      color: 'white',
    },
    ghost: {
      background: 'transparent',
      color: 'var(--color-text)',
      borderColor: 'var(--color-border)',
    },
    danger: {
      background: 'var(--color-danger)',
      color: 'white',
    },
  }

  const disabledStyles = (disabled || loading) ? {
    opacity: 0.6,
    cursor: 'not-allowed',
  } : {}

  const combinedStyles: React.CSSProperties = {
    ...baseStyles,
    ...sizeStyles[size],
    ...variantStyles[variant],
    ...disabledStyles,
    ...style,
  }

  return (
    <button
      {...props}
      disabled={disabled || loading}
      style={combinedStyles}
    >
      {loading ? (
        <>
          <span style={{ marginRight: '4px' }}>⟳</span>
          {children}
        </>
      ) : (
        children
      )}
    </button>
  )
}