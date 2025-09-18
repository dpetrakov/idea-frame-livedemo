// Страница создания инициативы
// TK-002/FE: Экран создания с формой, обработкой ошибок и успешным редиректом

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { CreateInitiativeForm } from '../features/initiatives/components/CreateInitiativeForm';
import { createInitiative } from '../features/initiatives/api';
import type { InitiativeCreate, ApiError } from '../features/initiatives/types';

export function CreateInitiative() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (data: InitiativeCreate) => {
    setLoading(true);
    setError(null);

    try {
      const newInitiative = await createInitiative(data);
      
      // Показываем уведомление об успехе (в будущем можно добавить toast)
      console.log('Инициатива создана:', newInitiative);
      
      // Перенаправляем на страницу созданной инициативы
      navigate(`/initiatives/${newInitiative.id}`);
      
    } catch (err) {
      console.error('Ошибка создания инициативы:', err);
      
      // Обработка API ошибок
      if (err && typeof err === 'object' && 'message' in err) {
        const apiError = err as ApiError;
        setError(apiError.message || 'Произошла ошибка при создании инициативы');
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Произошла неожиданная ошибка');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    navigate('/');
  };

  return (
    <div style={{
      maxWidth: '800px',
      margin: '0 auto',
      padding: 'var(--space-4)',
      minHeight: '100vh',
      display: 'flex',
      flexDirection: 'column'
    }}>
      {/* Заголовок страницы */}
      <header style={{ 
        marginBottom: 'var(--space-6)',
        paddingBottom: 'var(--space-4)',
        borderBottom: '1px solid var(--color-border)'
      }}>
        <h1 style={{ 
          margin: 0, 
          fontSize: 'var(--fs-2xl)', 
          fontWeight: 600,
          color: 'var(--color-text)',
          marginBottom: 'var(--space-2)'
        }}>
          Создание инициативы
        </h1>
        <p style={{
          margin: 0,
          color: 'var(--color-text-muted)',
          fontSize: 'var(--fs-md)',
          lineHeight: 1.5
        }}>
          Опишите вашу инициативу. Название должно быть кратким и понятным, 
          а в описании можно использовать Markdown для форматирования.
        </p>
      </header>

      {/* Основной контент */}
      <main style={{ flex: 1 }}>
        <CreateInitiativeForm
          onSubmit={handleSubmit}
          loading={loading}
          error={error}
        />
      </main>

      {/* Навигация */}
      <footer style={{
        marginTop: 'var(--space-6)',
        paddingTop: 'var(--space-4)',
        borderTop: '1px solid var(--color-border)',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center'
      }}>
        <button
          onClick={handleCancel}
          disabled={loading}
          style={{
            background: 'transparent',
            color: 'var(--color-text-muted)',
            border: '1px solid var(--color-border)',
            borderRadius: 'var(--radius)',
            padding: 'var(--space-2) var(--space-4)',
            cursor: loading ? 'not-allowed' : 'pointer',
            fontSize: 'var(--fs-md)',
            opacity: loading ? 0.6 : 1,
            transition: 'all 0.2s ease'
          }}
        >
          ← Назад к списку
        </button>
        
        <p style={{
          margin: 0,
          color: 'var(--color-text-muted)',
          fontSize: 'var(--fs-sm)'
        }}>
          Все поля с * обязательны для заполнения
        </p>
      </footer>
    </div>
  );
}