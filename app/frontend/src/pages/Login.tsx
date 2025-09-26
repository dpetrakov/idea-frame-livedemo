import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Input, Card } from '../shared/ui';
import { useAuth } from '../features/auth/useAuth';
import { ApiError } from '../shared/lib/api-client';

export function LoginPage() {
  const navigate = useNavigate();
  const { loginByEmailCode, requestEmailCode } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');

  const [formData, setFormData] = useState({
    email: '',
    emailCode: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    setSuccess('');

    try {
      await loginByEmailCode({ email: formData.email, emailCode: formData.emailCode });
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
      setSuccess('Код отправлен на указанный e‑mail');
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

        <div style={{ marginBottom: 'var(--space-6)' }} />

        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
          <Input
            label="Корпоративный e‑mail"
            type="email"
            required
            value={formData.email}
            onChange={handleInputChange('email')}
            placeholder="name@axenix.pro"
          />

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

          {success && (
            <div
              style={{
                padding: 'var(--space-3)',
                background: 'hsl(142, 76%, 95%)',
                borderRadius: 'var(--radius)',
                color: 'var(--color-success)',
                fontSize: 'var(--fs-sm)',
              }}
            >
              {success}
            </div>
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
            Войти
          </Button>
        </form>
      </Card>
    </div>
  );
}