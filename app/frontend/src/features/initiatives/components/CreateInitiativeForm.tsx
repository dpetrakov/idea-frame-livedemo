// Форма создания инициативы
// TK-002/FE: Форма с валидацией названия/описания, Markdown превью, русская локализация

import React, { useState, useEffect } from 'react';
import { Button } from '../../../shared/ui/Button';
import { Input } from '../../../shared/ui/Input';
import { Card } from '../../../shared/ui/Card';
import { renderMarkdown } from '../../../shared/lib/markdown';
import type { InitiativeFormData, InitiativeFormErrors } from '../types';
import { INITIATIVE_LIMITS } from '../types';
import { validateInitiativeCreate } from '../api';

interface CreateInitiativeFormProps {
  onSubmit: (data: { title: string; description?: string }) => void;
  loading?: boolean;
  error?: string | null;
}

export function CreateInitiativeForm({ onSubmit, loading = false, error }: CreateInitiativeFormProps) {
  const [formData, setFormData] = useState<InitiativeFormData>({
    title: '',
    description: '',
  });
  
  const [errors, setErrors] = useState<InitiativeFormErrors>({});
  const [showPreview, setShowPreview] = useState(false);
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  // Валидация в реальном времени
  useEffect(() => {
    const newErrors: InitiativeFormErrors = {};
    
    if (touched.title) {
      if (!formData.title.trim()) {
        newErrors.title = 'Название обязательно';
      } else if (formData.title.length > INITIATIVE_LIMITS.TITLE_MAX) {
        newErrors.title = `Название не должно превышать ${INITIATIVE_LIMITS.TITLE_MAX} символов`;
      }
    }
    
    if (touched.description) {
      if (formData.description.length > INITIATIVE_LIMITS.DESCRIPTION_MAX) {
        newErrors.description = `Описание не должно превышать ${INITIATIVE_LIMITS.DESCRIPTION_MAX} символов`;
      }
    }
    
    setErrors(newErrors);
  }, [formData, touched]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Отметить все поля как тронутые при отправке
    setTouched({ title: true, description: true });
    
    // Валидация API
    const apiErrors = validateInitiativeCreate({
      title: formData.title,
      description: formData.description || null,
    });
    
    if (apiErrors.length > 0) {
      setErrors({ title: apiErrors.join(', ') });
      return;
    }
    
    // Если валидация пройдена, отправляем данные
    onSubmit({
      title: formData.title.trim(),
      description: formData.description.trim() || undefined,
    });
  };

  const handleTitleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({ ...prev, title: e.target.value }));
  };

  const handleDescriptionChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setFormData(prev => ({ ...prev, description: e.target.value }));
  };

  const handleBlur = (field: string) => {
    setTouched(prev => ({ ...prev, [field]: true }));
  };

  const isFormValid = !Object.keys(errors).length && formData.title.trim().length > 0;
  const titleCharsLeft = INITIATIVE_LIMITS.TITLE_MAX - formData.title.length;
  const descriptionCharsLeft = INITIATIVE_LIMITS.DESCRIPTION_MAX - formData.description.length;

  return (
    <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
      {/* Общая ошибка */}
      {error && (
        <Card style={{ 
          background: 'var(--color-danger)', 
          color: 'white', 
          borderColor: 'var(--color-danger)' 
        }}>
          {error}
        </Card>
      )}

      {/* Поле названия */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
        <label htmlFor="title" style={{ 
          fontWeight: 600, 
          color: 'var(--color-text)',
          fontSize: 'var(--fs-sm)' 
        }}>
          Название инициативы *
        </label>
        <Input
          id="title"
          value={formData.title}
          onChange={handleTitleChange}
          onBlur={() => handleBlur('title')}
          placeholder="Введите название инициативы"
          disabled={loading}
          style={{ 
            borderColor: errors.title ? 'var(--color-danger)' : undefined,
            ...(loading && { opacity: 0.7 })
          }}
        />
        <div style={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          fontSize: 'var(--fs-xs)', 
          color: 'var(--color-text-muted)' 
        }}>
          <span style={{ color: errors.title ? 'var(--color-danger)' : undefined }}>
            {errors.title || ''}
          </span>
          <span style={{ color: titleCharsLeft < 20 ? 'var(--color-danger)' : undefined }}>
            {titleCharsLeft} символов осталось
          </span>
        </div>
      </div>

      {/* Поле описания */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <label htmlFor="description" style={{ 
            fontWeight: 600, 
            color: 'var(--color-text)',
            fontSize: 'var(--fs-sm)' 
          }}>
            Описание (опционально)
          </label>
          <div style={{ display: 'flex', gap: 'var(--space-2)' }}>
            <button
              type="button"
              onClick={() => setShowPreview(false)}
              disabled={loading}
              style={{
                background: !showPreview ? 'var(--color-highlight)' : 'transparent',
                color: !showPreview ? 'white' : 'var(--color-text)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--radius)',
                padding: 'var(--space-1) var(--space-2)',
                fontSize: 'var(--fs-xs)',
                cursor: 'pointer',
              }}
            >
              Редактор
            </button>
            <button
              type="button"
              onClick={() => setShowPreview(true)}
              disabled={loading}
              style={{
                background: showPreview ? 'var(--color-highlight)' : 'transparent',
                color: showPreview ? 'white' : 'var(--color-text)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--radius)',
                padding: 'var(--space-1) var(--space-2)',
                fontSize: 'var(--fs-xs)',
                cursor: 'pointer',
              }}
            >
              Превью
            </button>
          </div>
        </div>

        {!showPreview ? (
          <textarea
            id="description"
            value={formData.description}
            onChange={handleDescriptionChange}
            onBlur={() => handleBlur('description')}
            placeholder="Опишите инициативу. Поддерживается Markdown: **жирный**, *курсив*, [ссылки](url), списки и код."
            disabled={loading}
            rows={8}
            style={{
              width: '100%',
              padding: 'var(--space-3)',
              borderRadius: 'var(--radius)',
              border: `1px solid ${errors.description ? 'var(--color-danger)' : 'var(--color-border)'}`,
              background: 'white',
              resize: 'vertical',
              minHeight: '120px',
              fontFamily: 'var(--font-sans)',
              fontSize: 'var(--fs-md)',
              lineHeight: 1.5,
              ...(loading && { opacity: 0.7 })
            }}
          />
        ) : (
          <Card style={{ 
            minHeight: '120px', 
            padding: 'var(--space-3)',
            background: 'var(--color-bg-soft)',
            border: `1px solid var(--color-border)`,
            overflow: 'auto'
          }}>
            {formData.description.trim() ? (
              <div 
                dangerouslySetInnerHTML={{ __html: renderMarkdown(formData.description) }}
                style={{
                  lineHeight: 1.6,
                  fontSize: 'var(--fs-md)',
                  color: 'var(--color-text)'
                }}
              />
            ) : (
              <p style={{ 
                color: 'var(--color-text-muted)', 
                fontStyle: 'italic' 
              }}>
                Превью описания появится здесь...
              </p>
            )}
          </Card>
        )}

        <div style={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          fontSize: 'var(--fs-xs)', 
          color: 'var(--color-text-muted)' 
        }}>
          <span style={{ color: errors.description ? 'var(--color-danger)' : undefined }}>
            {errors.description || ''}
          </span>
          <span style={{ color: descriptionCharsLeft < 500 ? 'var(--color-danger)' : undefined }}>
            {descriptionCharsLeft} символов осталось
          </span>
        </div>
      </div>

      {/* Кнопка отправки */}
      <Button
        type="submit"
        variant="primary"
        disabled={!isFormValid || loading}
        style={{
          marginTop: 'var(--space-4)',
          ...((!isFormValid || loading) && { opacity: 0.6, cursor: 'not-allowed' })
        }}
      >
        {loading ? 'Создание...' : 'Создать инициативу'}
      </Button>
    </form>
  );
}