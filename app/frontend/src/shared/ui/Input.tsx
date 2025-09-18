import React from 'react'

export type InputProps = React.InputHTMLAttributes<HTMLInputElement> & {
  label?: string
  error?: string
  fullWidth?: boolean
}

export function Input({
  label,
  error,
  fullWidth = true,
  style,
  id,
  ...props
}: InputProps) {
  const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`

  const inputStyles: React.CSSProperties = {
    width: fullWidth ? '100%' : 'auto',
    padding: '12px',
    borderRadius: 'var(--radius)',
    border: error ? '1px solid var(--color-danger)' : '1px solid var(--color-border)',
    background: 'white',
    fontSize: 'var(--fs-md)',
    transition: 'border-color 0.2s ease',
    minHeight: '44px', // Минимум для touch targets
    ...style,
  }

  const labelStyles: React.CSSProperties = {
    display: 'block',
    marginBottom: 'var(--space-1)',
    fontSize: 'var(--fs-sm)',
    fontWeight: 500,
    color: 'var(--color-text)',
  }

  const errorStyles: React.CSSProperties = {
    display: 'block',
    marginTop: 'var(--space-1)',
    fontSize: 'var(--fs-xs)',
    color: 'var(--color-danger)',
  }

  const containerStyles: React.CSSProperties = {
    width: fullWidth ? '100%' : 'auto',
  }

  return (
    <div style={containerStyles}>
      {label && (
        <label htmlFor={inputId} style={labelStyles}>
          {label}
        </label>
      )}
      <input
        {...props}
        id={inputId}
        style={inputStyles}
      />
      {error && (
        <span style={errorStyles}>
          {error}
        </span>
      )}
    </div>
  )
}

export type TextareaProps = React.TextareaHTMLAttributes<HTMLTextAreaElement> & {
  label?: string
  error?: string
  fullWidth?: boolean
}

export function Textarea({
  label,
  error,
  fullWidth = true,
  style,
  id,
  ...props
}: TextareaProps) {
  const textareaId = id || `textarea-${Math.random().toString(36).substr(2, 9)}`

  const textareaStyles: React.CSSProperties = {
    width: fullWidth ? '100%' : 'auto',
    padding: '12px',
    borderRadius: 'var(--radius)',
    border: error ? '1px solid var(--color-danger)' : '1px solid var(--color-border)',
    background: 'white',
    fontSize: 'var(--fs-md)',
    transition: 'border-color 0.2s ease',
    resize: 'vertical',
    minHeight: '100px',
    ...style,
  }

  const labelStyles: React.CSSProperties = {
    display: 'block',
    marginBottom: 'var(--space-1)',
    fontSize: 'var(--fs-sm)',
    fontWeight: 500,
    color: 'var(--color-text)',
  }

  const errorStyles: React.CSSProperties = {
    display: 'block',
    marginTop: 'var(--space-1)',
    fontSize: 'var(--fs-xs)',
    color: 'var(--color-danger)',
  }

  const containerStyles: React.CSSProperties = {
    width: fullWidth ? '100%' : 'auto',
  }

  return (
    <div style={containerStyles}>
      {label && (
        <label htmlFor={textareaId} style={labelStyles}>
          {label}
        </label>
      )}
      <textarea
        {...props}
        id={textareaId}
        style={textareaStyles}
      />
      {error && (
        <span style={errorStyles}>
          {error}
        </span>
      )}
    </div>
  )
}