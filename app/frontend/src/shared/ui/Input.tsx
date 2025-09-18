import React from 'react';

type InputProps = React.InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  error?: string;
  fullWidth?: boolean;
};

export function Input({
  label,
  error,
  fullWidth = true,
  style,
  ...props
}: InputProps) {
  const inputStyles: React.CSSProperties = {
    width: fullWidth ? '100%' : 'auto',
    padding: '12px',
    borderRadius: 'var(--radius)',
    border: `1px solid ${error ? 'var(--color-danger)' : 'var(--color-border)'}`,
    background: 'white',
    fontSize: 'var(--fs-md)',
    transition: 'border-color 0.2s ease',
    ...style,
  };

  return (
    <div style={{ width: fullWidth ? '100%' : 'auto' }}>
      {label && (
        <label
          style={{
            display: 'block',
            marginBottom: '8px',
            fontSize: 'var(--fs-sm)',
            fontWeight: 500,
            color: 'var(--color-text)',
          }}
        >
          {label}
        </label>
      )}
      <input style={inputStyles} {...props} />
      {error && (
        <span
          style={{
            display: 'block',
            marginTop: '4px',
            fontSize: 'var(--fs-sm)',
            color: 'var(--color-danger)',
          }}
        >
          {error}
        </span>
      )}
    </div>
  );
}