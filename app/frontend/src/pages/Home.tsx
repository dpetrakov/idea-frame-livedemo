import { useAuth } from '../features/auth/useAuth';
import { Button, Card } from '../shared/ui';

export function HomePage() {
  const { user, logout } = useAuth();

  return (
    <div style={{ minHeight: '100vh', background: 'var(--color-bg-soft)' }}>
      <header
        style={{
          background: 'white',
          borderBottom: '1px solid var(--color-border)',
          padding: 'var(--space-4)',
        }}
      >
        <div
          style={{
            maxWidth: '1200px',
            margin: '0 auto',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <h1 style={{ color: 'var(--color-highlight)', fontSize: 'var(--fs-xl)' }}>
            Idea Frame
          </h1>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)' }}>
            <span>Привет, {user?.displayName}!</span>
            <Button variant="ghost" onClick={logout}>
              Выйти
            </Button>
          </div>
        </div>
      </header>

      <main style={{ padding: 'var(--space-6)', maxWidth: '1200px', margin: '0 auto' }}>
        <Card>
          <h2 style={{ marginBottom: 'var(--space-4)' }}>Список инициатив</h2>
          <p style={{ color: 'var(--color-text-muted)' }}>
            Здесь будет список инициатив. Функционал будет добавлен в следующих задачах.
          </p>
        </Card>
      </main>
    </div>
  );
}