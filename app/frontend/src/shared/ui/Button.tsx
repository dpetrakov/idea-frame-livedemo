import React from 'react';

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'ghost' | 'danger';
  fullWidth?: boolean;
  loading?: boolean;
};

export function Button({
  variant = 'primary',
  fullWidth = false,
  loading = false,
  style,
  children,
  disabled,
  ...props
}: ButtonProps) {
  const styles: React.CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
    padding: '10px 14px',
    borderRadius: 'var(--radius)',
    border: '1px solid transparent',
    boxShadow: 'var(--shadow-sm)',
    fontWeight: 600,
    fontSize: 'var(--fs-md)',
    transition: 'all 0.2s ease',
    width: fullWidth ? '100%' : 'auto',
    opacity: disabled || loading ? 0.6 : 1,
    cursor: disabled || loading ? 'not-allowed' : 'pointer',
    ...(variant === 'primary' && {
      background: 'var(--color-highlight)',
      color: 'white',
    }),
    ...(variant === 'ghost' && {
      background: 'transparent',
      color: 'var(--color-text)',
      borderColor: 'var(--color-border)',
    }),
    ...(variant === 'danger' && {
      background: 'var(--color-danger)',
      color: 'white',
    }),
    ...style,
  };

  return (
    <button
      style={styles}
      disabled={disabled || loading}
      {...props}
    >
      {loading ? 'Загрузка...' : children}
    </button>
  );
}