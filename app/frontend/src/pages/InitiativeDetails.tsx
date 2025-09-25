// Страница деталей инициативы
// TK-002/FE: Экран просмотра карточки инициативы с загрузкой по ID

import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { InitiativeCard } from '../features/initiatives/components/InitiativeCard';
import { getInitiativeById, updateInitiative, getComments, addComment, getUsers, deleteInitiative } from '../features/initiatives/api';
import { Card } from '../shared/ui/Card';
import { Button } from '../shared/ui/Button';
import type { Initiative, ApiError, LoadingState, InitiativeUpdate, UserBrief } from '../features/initiatives/types';
import { useAuth } from '../features/auth/useAuth';

export function InitiativeDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  
  const [initiative, setInitiative] = useState<Initiative | null>(null);
  const [metaEditing, setMetaEditing] = useState(false);
  const [draftTitle, setDraftTitle] = useState<string>('');
  const [draftDescription, setDraftDescription] = useState<string>('');
  const [metaSaving, setMetaSaving] = useState(false);
  const [metaError, setMetaError] = useState<string | null>(null);
  const [metaSaved, setMetaSaved] = useState(false);
  const [loadingState, setLoadingState] = useState<LoadingState>('idle');
  const [error, setError] = useState<string | null>(null);

  // Черновик значений атрибутов для редактирования
  const [draftValue, setDraftValue] = useState<number | null | undefined>(undefined);
  const [draftSpeed, setDraftSpeed] = useState<number | null | undefined>(undefined);
  const [draftCost, setDraftCost] = useState<number | null | undefined>(undefined);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  // TK-006: состояние для выбора ответственного
  const [users, setUsers] = useState<UserBrief[]>([]);
  const [usersLoading, setUsersLoading] = useState(false);
  const [usersError, setUsersError] = useState<string | null>(null);
  const [draftAssigneeId, setDraftAssigneeId] = useState<string>('');
  const [assigneeError, setAssigneeError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) {
      setError('ID инициативы не указан');
      setLoadingState('error');
      return;
    }

    loadInitiative(id);
    loadUsersList();
  }, [id]);

  // Скроллим к блоку редактирования, когда он становится видимым
  useEffect(() => {
    if (metaEditing) {
      const el = document.getElementById('meta-editor');
      if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  }, [metaEditing]);

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
      setDraftAssigneeId(data.assignee?.id ?? '');
      setDraftTitle(data.title);
      setDraftDescription(data.description ?? '');
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

  const loadUsersList = async () => {
    setUsersLoading(true);
    setUsersError(null);
    try {
      const list = await getUsers();
      setUsers(list);
    } catch (err) {
      const apiErr = err as ApiError;
      setUsersError(apiErr?.message || 'Не удалось загрузить пользователей');
    } finally {
      setUsersLoading(false);
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
    setMetaEditing(true);
    setMetaSaved(false);
    setMetaError(null);
  };

  const handleDelete = async () => {
    if (!id) return;
    const confirmed = window.confirm('Удалить инициативу? Это действие можно отменить только через БД.');
    if (!confirmed) return;
    try {
      await deleteInitiative(id);
      navigate('/');
    } catch (err) {
      const apiError = err as ApiError;
      alert(apiError?.message || 'Не удалось удалить инициативу');
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

  const isDirty = (
    (initiative?.value ?? null) !== (draftValue ?? null) ||
    (initiative?.speed ?? null) !== (draftSpeed ?? null) ||
    (initiative?.cost ?? null) !== (draftCost ?? null) ||
    ((initiative?.assignee?.id ?? '') !== draftAssigneeId)
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
    if ((initiative.assignee?.id ?? '') !== draftAssigneeId) payload.assigneeId = draftAssigneeId || null;

    try {
      const updated = await updateInitiative(id, payload);
      setInitiative(updated);
      setDraftValue(updated.value ?? null);
      setDraftSpeed(updated.speed ?? null);
      setDraftCost(updated.cost ?? null);
      setDraftAssigneeId(updated.assignee?.id ?? '');
      setAssigneeError(null);
      setSaved(true);
    } catch (err) {
      console.error('Ошибка сохранения инициативы:', err);
      if (err && typeof err === 'object' && 'message' in err) {
        const apiError = err as ApiError;
        setSaveError(apiError.message || 'Ошибка сохранения');
        // Inline ошибка под селектом для TK-006
        if (apiError.details && (apiError.details as any).assigneeId) {
          setAssigneeError((apiError.details as any).assigneeId);
        }
      } else if (err instanceof Error) {
        setSaveError(err.message);
      } else {
        setSaveError('Произошла неожиданная ошибка');
      }
    } finally {
      setSaving(false);
    }
  };

  const metaDirty = initiative && (initiative.title !== draftTitle || (initiative.description ?? '') !== draftDescription);

  const handleMetaSave = async () => {
    if (!initiative || !id) return;
    setMetaSaving(true);
    setMetaError(null);
    setMetaSaved(false);
    const payload: InitiativeUpdate = {};
    if (initiative.title !== draftTitle) payload.title = draftTitle;
    if ((initiative.description ?? '') !== draftDescription) payload.description = draftDescription;
    try {
      const updated = await updateInitiative(id, payload);
      setInitiative(updated);
      setDraftTitle(updated.title);
      setDraftDescription(updated.description ?? '');
      setMetaSaved(true);
      setMetaEditing(false);
    } catch (err) {
      if (err && typeof err === 'object' && 'message' in err) {
        const apiError = err as ApiError;
        setMetaError(apiError.message || 'Ошибка сохранения');
      } else if (err instanceof Error) {
        setMetaError(err.message);
      } else {
        setMetaError('Произошла неожиданная ошибка');
      }
    } finally {
      setMetaSaving(false);
    }
  };

  const handleVoteChange = (updatedInitiative: Initiative) => {
    setInitiative(updatedInitiative);
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
        justifyContent: 'flex-start',
        alignItems: 'center'
      }}>
        <Button variant="ghost" onClick={handleBack}>
          ← Назад к списку
        </Button>
      </header>

      {/* Основной контент */}
      <main>
        <InitiativeCard 
          initiative={initiative} 
          showFullDescription={true}
          onEdit={handleEdit}
          onDelete={user?.isAdmin ? handleDelete : undefined}
          canEditAdminFields={!!user?.isAdmin}
          onVoteChange={handleVoteChange}
        />

        {/* Редактирование названия и описания */}
        {metaEditing && (
          <section id="meta-editor" style={{ marginTop: 'var(--space-6)' }}>
            <Card style={{ padding: 'var(--space-4)' }}>
              <h3 style={{ marginTop: 0, marginBottom: 'var(--space-4)' }}>Редактирование</h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
                <div>
                  <div style={{ fontSize: 'var(--fs-sm)', color: 'var(--color-text-muted)', marginBottom: 'var(--space-1)' }}>Название</div>
                  <input
                    type="text"
                    value={draftTitle}
                    onChange={(e) => setDraftTitle(e.target.value)}
                    maxLength={140}
                    style={{ width: '100%', padding: '10px', borderRadius: 'var(--radius)', border: '1px solid var(--color-border)' }}
                  />
                </div>
                <div>
                  <div style={{ fontSize: 'var(--fs-sm)', color: 'var(--color-text-muted)', marginBottom: 'var(--space-1)' }}>Описание (Markdown)</div>
                  <textarea
                    value={draftDescription}
                    onChange={(e) => setDraftDescription(e.target.value)}
                    rows={10}
                    style={{ width: '100%', padding: '10px', borderRadius: 'var(--radius)', border: '1px solid var(--color-border)', fontFamily: 'inherit' }}
                  />
                </div>
                {metaError && (
                  <div style={{ color: 'var(--color-danger)' }}>{metaError}</div>
                )}
                {metaSaved && !metaError && (
                  <div style={{ color: 'var(--color-success)' }}>Сохранено</div>
                )}
                <div style={{ display: 'flex', gap: 'var(--space-3)' }}>
                  <Button
                    variant="primary"
                    onClick={handleMetaSave}
                    disabled={!metaDirty || metaSaving}
                  >
                    {metaSaving ? 'Сохранение...' : 'Сохранить изменения'}
                  </Button>
                  <Button
                    variant="ghost"
                    onClick={() => {
                      setDraftTitle(initiative.title);
                      setDraftDescription(initiative.description ?? '');
                      setMetaError(null);
                      setMetaSaved(false);
                      setMetaEditing(false);
                    }}
                  >
                    Отменить
                  </Button>
                </div>
              </div>
            </Card>
          </section>
        )}

        {/* TK-006: Выбор ответственного */}
        {user?.isAdmin && (
        <section id="assignee-editor" style={{ marginTop: 'var(--space-6)' }}>
          <Card style={{ padding: 'var(--space-4)' }}>
            <h3 style={{ marginTop: 0, marginBottom: 'var(--space-4)' }}>Ответственный</h3>
            <div>
              <label htmlFor="assignee-select" style={{ display: 'block', marginBottom: 'var(--space-2)', color: 'var(--color-text-muted)' }}>
                Назначить ответственного
              </label>
              <select
                id="assignee-select"
                value={draftAssigneeId}
                onChange={(e) => setDraftAssigneeId(e.target.value)}
                disabled={usersLoading || saving}
                style={{ width: '100%', padding: '10px', borderRadius: 'var(--radius)', border: '1px solid var(--color-border)' }}
              >
                <option value="">Не назначено</option>
                {users.map(u => (
                  <option key={u.id} value={u.id}>{u.displayName} ({u.login})</option>
                ))}
              </select>
              {usersError && (
                <div style={{ color: 'var(--color-danger)', marginTop: 'var(--space-2)' }}>{usersError}</div>
              )}
              {assigneeError && (
                <div style={{ color: 'var(--color-danger)', marginTop: 'var(--space-2)' }}>{assigneeError}</div>
              )}
            </div>
          </Card>
        </section>
        )}

        {/* Редактор атрибутов TK-003 */}
        {user?.isAdmin && (
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
                  setDraftAssigneeId(initiative.assignee?.id ?? '');
                  setAssigneeError(null);
                  setSaveError(null);
                  setSaved(false);
                }}
              >
                Сбросить изменения
              </Button>
            </div>
          </Card>
        </section>
        )}
      </main>

      {/* Комментарии TK-004 */}
      <section id="comments" style={{ marginTop: 'var(--space-6)' }}>
        <Card style={{ padding: 'var(--space-4)' }}>
          <CommentsSection initiativeId={id!} />
        </Card>
      </section>

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
          Связанные функции будут добавлены в следующих версиях
        </p>
      </footer>
    </div>
  );
}

function CommentsSection({ initiativeId }: { initiativeId: string }) {
  const [list, setList] = useState<import('../features/initiatives/types').CommentsList | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [text, setText] = useState('');
  const [sending, setSending] = useState(false);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getComments(initiativeId, { limit: 50, offset: 0 });
      setList(data);
    } catch (e) {
      const err = e as any;
      setError(err?.message || 'Не удалось загрузить комментарии');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [initiativeId]);

  const canSend = text.trim().length > 0 && text.trim().length <= 1000;

  const onSend = async () => {
    if (!canSend) return;
    setSending(true);
    try {
      const created = await addComment(initiativeId, text.trim());
      setText('');
      // Добавим созданный комментарий в конец списка без повторной загрузки
      setList(prev => prev ? { ...prev, items: [...prev.items, created], total: prev.total + 1 } : prev);
      // Скролл к низу можно реализовать через ref, опустим для простоты
    } catch (e) {
      const err = e as any;
      setError(err?.message || 'Не удалось отправить комментарий');
    } finally {
      setSending(false);
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      <h3 style={{ margin: 0 }}>Комментарии</h3>

      {/* Состояния загрузки/ошибки/пусто */}
      {loading && (
        <div style={{ color: 'var(--color-text-muted)' }}>Загрузка комментариев…</div>
      )}
      {!loading && error && (
        <div style={{ color: 'var(--color-danger)' }}>{error}</div>
      )}
      {!loading && !error && list && list.items.length === 0 && (
        <div style={{ color: 'var(--color-text-muted)' }}>Пока нет комментариев</div>
      )}

      {/* Лента */}
      {!loading && !error && list && list.items.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
          {list.items.map(item => (
            <div key={item.id} style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
              <div style={{ fontSize: 'var(--fs-sm)', color: 'var(--color-text-muted)' }}>{new Date(item.createdAt).toLocaleString('ru-RU')}</div>
              <div style={{
                alignSelf: 'flex-start',
                background: 'var(--color-bg-soft)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--radius)',
                padding: 'var(--space-3)',
                maxWidth: '90%'
              }}>
                <div style={{ fontWeight: 600, marginBottom: 6 }}>{item.author.displayName}</div>
                <div style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>{item.text}</div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Форма отправки */}
      <div style={{ display: 'flex', gap: 'var(--space-3)', alignItems: 'flex-start' }}>
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Напишите комментарий…"
          rows={2}
          style={{
            flex: 1,
            padding: '12px',
            borderRadius: 'var(--radius)',
            border: '1px solid var(--color-border)',
            resize: 'vertical'
          }}
        />
        <Button variant="primary" onClick={onSend} disabled={!canSend || sending}>
          {sending ? 'Отправка…' : 'Отправить'}
        </Button>
      </div>
    </div>
  );
}
