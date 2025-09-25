import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Card } from '../shared/ui';
import { useAuth } from '../features/auth/useAuth';
import { ApiError } from '../shared/lib/api-client';

export function LoginPage() {
  const navigate = useNavigate();
  const { login, register, requestEmailCode } = useAuth();
  const [isRegister, setIsRegister] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');

  const [formData, setFormData] = useState({
    login: '',
    displayName: '',
    email: '',
    emailCode: '',
    password: '',
    confirmPassword: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      if (isRegister) {
        await register({
          login: formData.login,
          displayName: formData.displayName,
          email: formData.email,
          emailCode: formData.emailCode,
          password: formData.password,
          confirmPassword: formData.confirmPassword,
        });
      } else {
        await login({
          login: formData.login,
          password: formData.password,
        });
      }
      navigate('/');
    } catch (err) {
      const apiError = err as ApiError;
      setError(apiError.message || 'Произошла ошибка');
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({ ...prev, [field]: e.target.value }));
    setError('');
  };

  const handleRequestCode = async () => {
    setError('');
    if (!formData.email) {
      setError('Укажите e‑mail');
      return;
    }
    try {
      setLoading(true);
      await requestEmailCode(formData.email);
      setError('Код отправлен на указанный e‑mail');
    } catch (err) {
      const apiError = err as ApiError;
      setError(apiError.message || 'Не удалось отправить код');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'var(--color-bg-soft)',
        padding: 'var(--space-4)',
      }}
    >
      <Card style={{ width: '100%', maxWidth: '400px' }}>
        <div style={{ textAlign: 'center', marginBottom: 'var(--space-6)' }}>
          <h1 style={{ color: 'var(--color-highlight)', marginBottom: 'var(--space-2)' }}>
            MeetAx Next
          </h1>
          <p style={{ color: 'var(--color-text-muted)' }}>
            Площадка, где рождаются и отбираются идеи для будущего развития MeetAx
          </p>
        </div>

        <div
          style={{
            display: 'flex',
            gap: 'var(--space-2)',
            marginBottom: 'var(--space-6)',
            borderBottom: '1px solid var(--color-border)',
          }}
        >
          <button
            onClick={() => setIsRegister(false)}
            style={{
              flex: 1,
              background: 'transparent',
              border: 'none',
              padding: 'var(--space-3)',
              fontSize: 'var(--fs-md)',
              fontWeight: 600,
              color: !isRegister ? 'var(--color-highlight)' : 'var(--color-text-muted)',
              borderBottom: !isRegister ? '2px solid var(--color-highlight)' : '2px solid transparent',
              cursor: 'pointer',
              transition: 'all 0.2s ease',
            }}
          >
            Вход
          </button>
          <button
            onClick={() => setIsRegister(true)}
            style={{
              flex: 1,
              background: 'transparent',
              border: 'none',
              padding: 'var(--space-3)',
              fontSize: 'var(--fs-md)',
              fontWeight: 600,
              color: isRegister ? 'var(--color-highlight)' : 'var(--color-text-muted)',
              borderBottom: isRegister ? '2px solid var(--color-highlight)' : '2px solid transparent',
              cursor: 'pointer',
              transition: 'all 0.2s ease',
            }}
          >
            Регистрация
          </button>
        </div>

        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
          <Input
            label="Логин"
            type="text"
            required
            value={formData.login}
            onChange={handleInputChange('login')}
            placeholder="Введите логин"
          />

          {isRegister && (
            <Input
              label="Отображаемое имя"
              type="text"
              required
              value={formData.displayName}
              onChange={handleInputChange('displayName')}
              placeholder="Введите ваше имя"
            />
          )}

          {isRegister && (
            <Input
              label="Корпоративный e‑mail"
              type="email"
              required
              value={formData.email}
              onChange={handleInputChange('email')}
              placeholder="name@axenix.pro"
            />
          )}

          {isRegister && (
            <div style={{ display: 'flex', gap: 'var(--space-2)' }}>
              <div style={{ flex: 1 }}>
                <Input
                  label="Код из письма"
                  type="text"
                  required
                  value={formData.emailCode}
                  onChange={handleInputChange('emailCode')}
                  placeholder="6 цифр"
                />
              </div>
              <div style={{ alignSelf: 'flex-end' }}>
                <Button type="button" onClick={handleRequestCode} loading={loading}>
                  Получить код
                </Button>
              </div>
            </div>
          )}

          <Input
            label="Пароль"
            type="password"
            required
            value={formData.password}
            onChange={handleInputChange('password')}
            placeholder="Введите пароль"
          />

          {isRegister && (
            <Input
              label="Подтвердите пароль"
              type="password"
              required
              value={formData.confirmPassword}
              onChange={handleInputChange('confirmPassword')}
              placeholder="Повторите пароль"
            />
          )}

          {error && (
            <div
              style={{
                padding: 'var(--space-3)',
                background: 'hsl(0, 100%, 95%)',
                borderRadius: 'var(--radius)',
                color: 'var(--color-danger)',
                fontSize: 'var(--fs-sm)',
              }}
            >
              {error}
            </div>
          )}

          <Button type="submit" fullWidth loading={loading}>
            {isRegister ? 'Зарегистрироваться' : 'Войти'}
          </Button>
        </form>
      </Card>
    </div>
  );
}