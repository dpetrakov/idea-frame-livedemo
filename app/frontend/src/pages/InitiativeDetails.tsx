// Страница деталей инициативы
// TK-002/FE: Экран просмотра карточки инициативы с загрузкой по ID

import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { InitiativeCard } from '../features/initiatives/components/InitiativeCard';
import { getInitiativeById, updateInitiative } from '../features/initiatives/api';
import { Card } from '../shared/ui/Card';
import { Button } from '../shared/ui/Button';
import type { Initiative, ApiError, LoadingState, InitiativeUpdate } from '../features/initiatives/types';

export function InitiativeDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  
  const [initiative, setInitiative] = useState<Initiative | null>(null);
  const [loadingState, setLoadingState] = useState<LoadingState>('idle');
  const [error, setError] = useState<string | null>(null);

  // Черновик значений атрибутов для редактирования
  const [draftValue, setDraftValue] = useState<number | null | undefined>(undefined);
  const [draftSpeed, setDraftSpeed] = useState<number | null | undefined>(undefined);
  const [draftCost, setDraftCost] = useState<number | null | undefined>(undefined);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

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
      // Инициализируем черновик текущими значениями
      setDraftValue(data.value ?? null);
      setDraftSpeed(data.speed ?? null);
      setDraftCost(data.cost ?? null);
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
    // Редактирование реализовано ниже как форма атрибутов
    const el = document.getElementById('attributes-editor');
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
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

  const isDirty = (
    (initiative?.value ?? null) !== (draftValue ?? null) ||
    (initiative?.speed ?? null) !== (draftSpeed ?? null) ||
    (initiative?.cost ?? null) !== (draftCost ?? null)
  );

  const handleSave = async () => {
    if (!initiative || !id) return;
    setSaving(true);
    setSaveError(null);
    setSaved(false);

    const payload: InitiativeUpdate = {};
    if ((initiative.value ?? null) !== (draftValue ?? null)) payload.value = draftValue ?? null;
    if ((initiative.speed ?? null) !== (draftSpeed ?? null)) payload.speed = draftSpeed ?? null;
    if ((initiative.cost ?? null) !== (draftCost ?? null)) payload.cost = draftCost ?? null;

    try {
      const updated = await updateInitiative(id, payload);
      setInitiative(updated);
      setDraftValue(updated.value ?? null);
      setDraftSpeed(updated.speed ?? null);
      setDraftCost(updated.cost ?? null);
      setSaved(true);
    } catch (err) {
      console.error('Ошибка сохранения инициативы:', err);
      if (err && typeof err === 'object' && 'message' in err) {
        const apiError = err as ApiError;
        setSaveError(apiError.message || 'Ошибка сохранения');
      } else if (err instanceof Error) {
        setSaveError(err.message);
      } else {
        setSaveError('Произошла неожиданная ошибка');
      }
    } finally {
      setSaving(false);
    }
  };

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

        {/* Редактор атрибутов TK-003 */}
        <section id="attributes-editor" style={{ marginTop: 'var(--space-6)' }}>
          <Card style={{ padding: 'var(--space-4)' }}>
            <h3 style={{ marginTop: 0, marginBottom: 'var(--space-4)' }}>Оценка инициативы</h3>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 'var(--space-4)' }}>
              <div>
                <div style={{ fontSize: 'var(--fs-sm)', color: 'var(--color-text-muted)', marginBottom: 'var(--space-1)' }}>Ценность</div>
                <select
                  value={draftValue ?? ''}
                  onChange={(e) => setDraftValue(e.target.value === '' ? null : Number(e.target.value))}
                  style={{ width: '100%', padding: '10px', borderRadius: 'var(--radius)', border: '1px solid var(--color-border)' }}
                >
                  <option value="">—</option>
                  {[1,2,3,4,5].map(v => <option key={v} value={v}>{v}</option>)}
                </select>
              </div>
              <div>
                <div style={{ fontSize: 'var(--fs-sm)', color: 'var(--color-text-muted)', marginBottom: 'var(--space-1)' }}>Скорость</div>
                <select
                  value={draftSpeed ?? ''}
                  onChange={(e) => setDraftSpeed(e.target.value === '' ? null : Number(e.target.value))}
                  style={{ width: '100%', padding: '10px', borderRadius: 'var(--radius)', border: '1px solid var(--color-border)' }}
                >
                  <option value="">—</option>
                  {[1,2,3,4,5].map(v => <option key={v} value={v}>{v}</option>)}
                </select>
              </div>
              <div>
                <div style={{ fontSize: 'var(--fs-sm)', color: 'var(--color-text-muted)', marginBottom: 'var(--space-1)' }}>Стоимость</div>
                <select
                  value={draftCost ?? ''}
                  onChange={(e) => setDraftCost(e.target.value === '' ? null : Number(e.target.value))}
                  style={{ width: '100%', padding: '10px', borderRadius: 'var(--radius)', border: '1px solid var(--color-border)' }}
                >
                  <option value="">—</option>
                  {[1,2,3,4,5].map(v => <option key={v} value={v}>{v}</option>)}
                </select>
              </div>
            </div>

            {saveError && (
              <div style={{ color: 'var(--color-danger)', marginTop: 'var(--space-3)' }}>{saveError}</div>
            )}
            {saved && !saveError && (
              <div style={{ color: 'var(--color-success)', marginTop: 'var(--space-3)' }}>Сохранено</div>
            )}

            <div style={{ display: 'flex', gap: 'var(--space-3)', marginTop: 'var(--space-4)' }}>
              <Button 
                variant="primary" 
                onClick={handleSave} 
                disabled={!isDirty || saving}
              >
                {saving ? 'Сохранение...' : 'Сохранить'}
              </Button>
              <Button 
                variant="ghost" 
                onClick={() => {
                  setDraftValue(initiative.value ?? null);
                  setDraftSpeed(initiative.speed ?? null);
                  setDraftCost(initiative.cost ?? null);
                  setSaveError(null);
                  setSaved(false);
                }}
              >
                Сбросить изменения
              </Button>
            </div>
          </Card>
        </section>
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