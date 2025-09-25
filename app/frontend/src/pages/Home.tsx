import { useEffect, useMemo, useState } from 'react';
import { useAuth } from '../features/auth/useAuth';
import { Button, Card } from '../shared/ui';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { getInitiativesList } from '../features/initiatives/api';
import type { Initiative, InitiativesList } from '../features/initiatives/types';
import { InitiativeCard } from '../features/initiatives/components/InitiativeCard';

const FILTERS = [
  { key: 'all', label: 'Все' },
  { key: 'mineCreated', label: 'Я предложил' },
  { key: 'assignedToMe', label: 'Назначено на меня' },
] as const;

const SORTS = [
  { key: 'weight', label: 'По весу' },
  { key: 'votes', label: 'По голосам' },
] as const;

type FilterKey = typeof FILTERS[number]['key'];
type SortKey = typeof SORTS[number]['key'];

export function HomePage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const currentFilter = useMemo<FilterKey>(() => {
    const f = searchParams.get('filter') as FilterKey | null;
    return f && (FILTERS as any).some((x: any) => x.key === f) ? f : 'all';
  }, [searchParams]);

  const currentSort = useMemo<SortKey>(() => {
    const s = searchParams.get('sort') as SortKey | null;
    return s && (SORTS as any).some((x: any) => x.key === s) ? s : 'weight';
  }, [searchParams]);

  const [items, setItems] = useState<Initiative[]>([]);
  const [total, setTotal] = useState(0);
  const [limit] = useState(20);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleCreateInitiative = () => {
    navigate('/initiatives/new');
  };

  // Загрузка первой страницы при смене фильтра или сортировки
  useEffect(() => {
    let cancelled = false;
    async function loadFirstPage() {
      setLoading(true);
      setError(null);
      setItems([]);
      setOffset(0);
      try {
        const res: InitiativesList = await getInitiativesList({ 
          filter: currentFilter, 
          sort: currentSort,
          limit: limit, 
          offset: 0 
        });
        if (cancelled) return;
        setItems(res.items);
        setTotal(res.total);
        setOffset(res.items.length);
      } catch (e: any) {
        if (cancelled) return;
        setError(e?.message || 'Не удалось загрузить список инициатив');
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    loadFirstPage();
    return () => { cancelled = true; };
  }, [currentFilter, currentSort, limit]);

  // Догрузка следующей страницы
  const loadMore = async () => {
    if (loadingMore) return;
    setLoadingMore(true);
    setError(null);
    try {
      const res = await getInitiativesList({ 
        filter: currentFilter, 
        sort: currentSort, 
        limit, 
        offset 
      });
      setItems(prev => [...prev, ...res.items]);
      setOffset(prev => prev + res.items.length);
      setTotal(res.total);
    } catch (e: any) {
      setError(e?.message || 'Ошибка загрузки следующей страницы');
    } finally {
      setLoadingMore(false);
    }
  };

  const onChangeFilter = (key: FilterKey) => {
    setSearchParams(prev => {
      const sp = new URLSearchParams(prev);
      if (key === 'all') sp.delete('filter'); else sp.set('filter', key);
      // Сбрасываем пагинацию
      return sp;
    });
  };

  const onChangeSort = (key: SortKey) => {
    setSearchParams(prev => {
      const sp = new URLSearchParams(prev);
      if (key === 'weight') sp.delete('sort'); else sp.set('sort', key);
      // Сбрасываем пагинацию
      return sp;
    });
  };

  const handleVoteChange = (updatedInitiative: Initiative) => {
    setItems(prev => prev.map(item => 
      item.id === updatedInitiative.id ? updatedInitiative : item
    ));
  };

  const hasMore = items.length < total;

  return (
    <div style={{ minHeight: '100vh', background: 'var(--color-bg-soft)' }}>
      <header
        style={{
          background: 'white',
          borderBottom: '1px solid var(--color-border)',
          padding: 'var(--space-4)',
          position: 'sticky',
          top: 0,
          zIndex: 10,
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
          <h1 style={{ color: 'var(--color-highlight)', fontSize: 'var(--fs-xl)', margin: 0 }}>
            MeetAx Next
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
          marginBottom: 'var(--space-4)'
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

        {/* Табы фильтров и сортировки */}
        <div style={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center',
          marginBottom: 'var(--space-4)',
          flexWrap: 'wrap',
          gap: 'var(--space-4)'
        }}>
          {/* Фильтры */}
          <div style={{ display: 'flex', gap: 'var(--space-2)' }}>
            {FILTERS.map(f => (
              <button
                key={f.key}
                onClick={() => onChangeFilter(f.key)}
                style={{
                  background: currentFilter === f.key ? 'var(--color-highlight)' : 'transparent',
                  color: currentFilter === f.key ? 'white' : 'var(--color-text)',
                  border: '1px solid var(--color-border)',
                  borderRadius: 'var(--radius)',
                  padding: '8px 12px',
                  fontWeight: 600,
                  cursor: 'pointer',
                }}
              >
                {f.label}
              </button>
            ))}
          </div>

          {/* Сортировка */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
            <span style={{ fontSize: 'var(--fs-sm)', color: 'var(--color-text-muted)' }}>
              Сортировка:
            </span>
            {SORTS.map(s => (
              <button
                key={s.key}
                onClick={() => onChangeSort(s.key)}
                style={{
                  background: currentSort === s.key ? 'var(--color-highlight)' : 'transparent',
                  color: currentSort === s.key ? 'white' : 'var(--color-text)',
                  border: '1px solid var(--color-border)',
                  borderRadius: 'var(--radius)',
                  padding: '6px 12px',
                  fontSize: 'var(--fs-sm)',
                  fontWeight: 500,
                  cursor: 'pointer',
                }}
              >
                {s.label}
              </button>
            ))}
          </div>
        </div>
        
        {/* Состояние ошибки */}
        {error && (
          <Card style={{ borderColor: 'var(--color-danger)', color: 'var(--color-danger)', padding: 'var(--space-3)', marginBottom: 'var(--space-4)' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 'var(--space-4)' }}>
              <span>{error}</span>
              <Button variant="ghost" onClick={() => { setSearchParams(prev => new URLSearchParams(prev)); }}>
                Повторить
              </Button>
            </div>
          </Card>
        )}

        {/* Скелетоны загрузки (первая загрузка) */}
        {loading && (
          <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: 'var(--space-4)' }}>
            {Array.from({ length: 6 }).map((_, i) => (
              <Card key={i} style={{ padding: 'var(--space-4)' }}>
                <div style={{ height: '20px', width: '60%', background: 'var(--color-bg-soft)', borderRadius: '8px', marginBottom: '12px' }} />
                <div style={{ height: '12px', width: '40%', background: 'var(--color-bg-soft)', borderRadius: '8px', marginBottom: '8px' }} />
                <div style={{ height: '12px', width: '80%', background: 'var(--color-bg-soft)', borderRadius: '8px' }} />
              </Card>
            ))}
          </div>
        )}

        {/* Пусто */}
        {!loading && items.length === 0 && (
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
                Создайте первую инициативу
              </h3>
              <p style={{ 
                margin: '0 auto var(--space-4)',
                color: 'var(--color-text-muted)',
                fontSize: 'var(--fs-md)',
                maxWidth: '500px'
              }}>
                Список фильтруется и сортируется по весу (убывание), затем по дате. Нажмите, чтобы создать новую инициативу.
              </p>
              <Button variant="primary" onClick={handleCreateInitiative}>
                Создать инициативу
              </Button>
            </div>
          </Card>
        )}

        {/* Список */}
        {!loading && items.length > 0 && (
          <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: 'var(--space-4)' }}>
            {items.map(item => (
              <div 
                key={item.id} 
                style={{ position: 'relative' }}
                onClick={(e) => {
                  // Не переходить если клик был на кнопках голосования
                  if ((e.target as HTMLElement).closest('[data-vote-buttons]')) {
                    return;
                  }
                  navigate(`/initiatives/${item.id}`);
                }}
              >
                <div style={{ cursor: 'pointer' }}>
                  <InitiativeCard 
                    initiative={item} 
                    showFullDescription={false} 
                    onVoteChange={handleVoteChange}
                  />
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Кнопка Показать ещё */}
        {!loading && hasMore && (
          <div style={{ display: 'flex', justifyContent: 'center', marginTop: 'var(--space-6)' }}>
            <Button variant="ghost" onClick={loadMore} disabled={loadingMore}>
              {loadingMore ? 'Загрузка...' : 'Показать ещё'}
            </Button>
          </div>
        )}
      </main>
    </div>
  );
}
