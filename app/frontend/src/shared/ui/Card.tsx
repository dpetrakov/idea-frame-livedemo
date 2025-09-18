import React from 'react'

export type CardProps = React.HTMLAttributes<HTMLDivElement> & {
  variant?: 'default' | 'elevated'
}

export function Card({ variant = 'default', children, style, ...props }: CardProps) {
  const baseStyles: React.CSSProperties = {
    background: 'var(--color-bg)',
    border: '1px solid var(--color-border)',
    borderRadius: 'var(--radius-lg)',
    padding: 'var(--space-4)',
  }

  const variantStyles = {
    default: {
      boxShadow: 'var(--shadow-sm)',
    },
    elevated: {
      boxShadow: 'var(--shadow-md)',
    },
  }

  const combinedStyles: React.CSSProperties = {
    ...baseStyles,
    ...variantStyles[variant],
    ...style,
  }

  return (
    <div {...props} style={combinedStyles}>
      {children}
    </div>
  )
}

export function CardHeader({ children, style, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  const headerStyles: React.CSSProperties = {
    marginBottom: 'var(--space-4)',
    paddingBottom: 'var(--space-3)',
    borderBottom: '1px solid var(--color-border)',
    ...style,
  }

  return (
    <div {...props} style={headerStyles}>
      {children}
    </div>
  )
}

export function CardContent({ children, style, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div {...props} style={style}>
      {children}
    </div>
  )
}

export function CardFooter({ children, style, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  const footerStyles: React.CSSProperties = {
    marginTop: 'var(--space-4)',
    paddingTop: 'var(--space-3)',
    borderTop: '1px solid var(--color-border)',
    display: 'flex',
    gap: 'var(--space-2)',
    ...style,
  }

  return (
    <div {...props} style={footerStyles}>
      {children}
    </div>
  )
}