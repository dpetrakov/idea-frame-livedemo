import { useAuth } from '../features/auth/useAuth';
import { Button, Card } from '../shared/ui';
import { useNavigate } from 'react-router-dom';

export function HomePage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  
  const handleCreateInitiative = () => {
    navigate('/initiatives/new');
  };

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
        {/* Заголовок и действие */}
        <div style={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center', 
          marginBottom: 'var(--space-6)' 
        }}>
          <div>
            <h2 style={{ margin: 0, fontSize: 'var(--fs-2xl)', marginBottom: 'var(--space-2)' }}>
              Инициативы
            </h2>
            <p style={{ margin: 0, color: 'var(--color-text-muted)', fontSize: 'var(--fs-md)' }}>
              Управление портфелем инициатив с оценкой приоритетов
            </p>
          </div>
          <Button variant="primary" onClick={handleCreateInitiative}>
            + Создать инициативу
          </Button>
        </div>
        
        {/* Основной контент */}
        <Card>
          <div style={{ textAlign: 'center', padding: 'var(--space-6)' }}>
            <div style={{ 
              fontSize: '48px', 
              marginBottom: 'var(--space-4)',
              color: 'var(--color-text-muted)' 
            }}>
              📋
            </div>
            <h3 style={{ 
              margin: 0, 
              marginBottom: 'var(--space-2)', 
              fontSize: 'var(--fs-xl)', 
              color: 'var(--color-text)' 
            }}>
              Начните с создания первой инициативы
            </h3>
            <p style={{ 
              margin: '0 auto var(--space-4)',
              color: 'var(--color-text-muted)',
              fontSize: 'var(--fs-md)',
              maxWidth: '500px'
            }}>
              Создавайте, оценивайте и управляйте инициативами. 
              Список инициатив с фильтрацией и сортировкой будет добавлен в следующих версиях.
            </p>
            <Button variant="primary" onClick={handleCreateInitiative}>
              Создать первую инициативу
            </Button>
          </div>
        </Card>
      </main>
    </div>
  );
}