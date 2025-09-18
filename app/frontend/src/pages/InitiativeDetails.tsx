// Страница деталей инициативы
// TK-002/FE: Экран просмотра карточки инициативы с загрузкой по ID

import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { InitiativeCard } from '../features/initiatives/components/InitiativeCard';
import { getInitiativeById } from '../features/initiatives/api';
import { Card } from '../shared/ui/Card';
import { Button } from '../shared/ui/Button';
import type { Initiative, ApiError, LoadingState } from '../features/initiatives/types';

export function InitiativeDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  
  const [initiative, setInitiative] = useState<Initiative | null>(null);
  const [loadingState, setLoadingState] = useState<LoadingState>('idle');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) {
      setError('ID инициативы не указан');
      setLoadingState('error');
      return;
    }

    loadInitiative(id);
  }, [id]);

  const loadInitiative = async (initiativeId: string) => {
    setLoadingState('loading');
    setError(null);

    try {
      const data = await getInitiativeById(initiativeId);
      setInitiative(data);
      setLoadingState('success');
    } catch (err) {
      console.error('Ошибка загрузки инициативы:', err);
      
      // Обработка API ошибок
      if (err && typeof err === 'object' && 'message' in err) {
        const apiError = err as ApiError;
        if (apiError.code === 'NOT_FOUND') {
          setError('Инициатива не найдена');
        } else if (apiError.code === 'UNAUTHORIZED') {
          setError('Требуется авторизация для просмотра инициативы');
        } else {
          setError(apiError.message || 'Ошибка загрузки инициативы');
        }
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Произошла неожиданная ошибка');
      }
      
      setLoadingState('error');
    }
  };

  const handleRetry = () => {
    if (id) {
      loadInitiative(id);
    }
  };

  const handleBack = () => {
    navigate('/');
  };

  const handleEdit = () => {
    if (initiative) {
      // В будущих задачах (TK-003, TK-006) будет реализовано редактирование
      console.log('Редактирование инициативы:', initiative.id);
      alert('Редактирование будет реализовано в следующих версиях');
    }
  };

  // Состояние загрузки
  if (loadingState === 'loading') {
    return (
      <div style={{
        maxWidth: '800px',
        margin: '0 auto',
        padding: 'var(--space-4)',
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center'
      }}>
        <Card style={{ 
          textAlign: 'center',
          padding: 'var(--space-6)'
        }}>
          <div style={{ 
            marginBottom: 'var(--space-4)',
            fontSize: 'var(--fs-lg)',
            color: 'var(--color-text-muted)'
          }}>
            Загружается...
          </div>
          <div style={{
            width: '40px',
            height: '40px',
            border: '3px solid var(--color-border)',
            borderTop: '3px solid var(--color-highlight)',
            borderRadius: '50%',
            animation: 'spin 1s linear infinite',
            margin: '0 auto'
          }} />
          <style>
            {`
              @keyframes spin {
                0% { transform: rotate(0deg); }
                100% { transform: rotate(360deg); }
              }
            `}
          </style>
        </Card>
      </div>
    );
  }

  // Состояние ошибки
  if (loadingState === 'error') {
    return (
      <div style={{
        maxWidth: '800px',
        margin: '0 auto',
        padding: 'var(--space-4)',
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center'
      }}>
        <Card style={{ 
          textAlign: 'center',
          padding: 'var(--space-6)',
          borderColor: 'var(--color-danger)'
        }}>
          <h2 style={{ 
            margin: 0, 
            marginBottom: 'var(--space-4)',
            color: 'var(--color-danger)',
            fontSize: 'var(--fs-xl)'
          }}>
            Ошибка загрузки
          </h2>
          <p style={{ 
            margin: 0, 
            marginBottom: 'var(--space-4)',
            color: 'var(--color-text-muted)',
            fontSize: 'var(--fs-md)'
          }}>
            {error}
          </p>
          <div style={{ 
            display: 'flex', 
            gap: 'var(--space-4)', 
            justifyContent: 'center',
            flexWrap: 'wrap'
          }}>
            <Button variant="primary" onClick={handleRetry}>
              Попробовать снова
            </Button>
            <Button variant="ghost" onClick={handleBack}>
              ← Назад к списку
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  // Состояние успеха - показываем инициативу
  if (!initiative) {
    return null;
  }

  return (
    <div style={{
      maxWidth: '800px',
      margin: '0 auto',
      padding: 'var(--space-4)',
      minHeight: '100vh'
    }}>
      {/* Навигация */}
      <header style={{ 
        marginBottom: 'var(--space-6)',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center'
      }}>
        <Button variant="ghost" onClick={handleBack}>
          ← Назад к списку
        </Button>
        
        <div style={{ display: 'flex', gap: 'var(--space-2)' }}>
          <Button variant="ghost" onClick={handleEdit}>
            ✏️ Редактировать
          </Button>
        </div>
      </header>

      {/* Основной контент */}
      <main>
        <InitiativeCard 
          initiative={initiative} 
          showFullDescription={true}
        />
      </main>

      {/* Дополнительные действия */}
      <footer style={{
        marginTop: 'var(--space-6)',
        paddingTop: 'var(--space-4)',
        borderTop: '1px solid var(--color-border)',
        display: 'flex',
        justifyContent: 'center'
      }}>
        <p style={{
          margin: 0,
          color: 'var(--color-text-muted)',
          fontSize: 'var(--fs-sm)',
          textAlign: 'center'
        }}>
          Комментарии и дополнительные возможности будут добавлены в следующих версиях
        </p>
      </footer>
    </div>
  );
}