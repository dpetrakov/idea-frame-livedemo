// Карточка инициативы
// TK-002/FE: Отображение деталей инициативы с автором, датами, весом и рендером Markdown

import React from 'react';
import { Card } from '../../../shared/ui/Card';
import { renderMarkdown } from '../../../shared/lib/markdown';
import type { Initiative } from '../types';

interface InitiativeCardProps {
  initiative: Initiative;
  showFullDescription?: boolean;
  onEdit?: () => void;
}

export function InitiativeCard({ initiative, showFullDescription = true, onEdit }: InitiativeCardProps) {
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const formatWeight = (weight: number) => {
    return weight.toFixed(2);
  };

  const formatAttribute = (value: number | null | undefined) => {
    return value !== null && value !== undefined ? value.toString() : '—';
  };

  return (
    <Card style={{ 
      display: 'flex', 
      flexDirection: 'column', 
      gap: 'var(--space-4)',
      position: 'relative'
    }}>
      {/* Заголовок */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 'var(--space-4)' }}>
        <h1 style={{ 
          margin: 0, 
          fontSize: 'var(--fs-2xl)', 
          fontWeight: 600,
          color: 'var(--color-text)',
          lineHeight: 1.3,
          flex: 1
        }}>
          {initiative.title}
        </h1>
        
        {/* Вес инициативы */}
        <div style={{
          background: 'var(--color-highlight)',
          color: 'white',
          padding: 'var(--space-2) var(--space-3)',
          borderRadius: 'var(--radius)',
          fontWeight: 600,
          fontSize: 'var(--fs-lg)',
          minWidth: '60px',
          textAlign: 'center',
          boxShadow: 'var(--shadow-sm)'
        }}>
          {formatWeight(initiative.weight)}
        </div>
      </div>

      {/* Мета-информация */}
      <div style={{ 
        display: 'flex', 
        flexDirection: 'column', 
        gap: 'var(--space-2)',
        padding: 'var(--space-3)',
        background: 'var(--color-bg-soft)',
        borderRadius: 'var(--radius)',
        fontSize: 'var(--fs-sm)'
      }}>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-4)' }}>
          {/* Автор */}
          <div>
            <span style={{ color: 'var(--color-text-muted)', fontWeight: 500 }}>Автор: </span>
            <span style={{ color: 'var(--color-text)', fontWeight: 600 }}>{initiative.author.displayName}</span>
          </div>
          
          {/* Ответственный */}
          <div>
            <span style={{ color: 'var(--color-text-muted)', fontWeight: 500 }}>Ответственный: </span>
            <span style={{ color: 'var(--color-text)', fontWeight: 600 }}>
              {initiative.assignee ? initiative.assignee.displayName : 'Не назначен'}
            </span>
          </div>
          
          {/* Комментарии */}
          <div>
            <span style={{ color: 'var(--color-text-muted)', fontWeight: 500 }}>Комментарии: </span>
            <span style={{ color: 'var(--color-text)', fontWeight: 600 }}>{initiative.commentsCount}</span>
          </div>
        </div>
        
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 'var(--space-4)' }}>
          {/* Создано */}
          <div>
            <span style={{ color: 'var(--color-text-muted)', fontWeight: 500 }}>Создано: </span>
            <span style={{ color: 'var(--color-text)' }}>{formatDate(initiative.createdAt)}</span>
          </div>
          
          {/* Обновлено */}
          {initiative.updatedAt !== initiative.createdAt && (
            <div>
              <span style={{ color: 'var(--color-text-muted)', fontWeight: 500 }}>Обновлено: </span>
              <span style={{ color: 'var(--color-text)' }}>{formatDate(initiative.updatedAt)}</span>
            </div>
          )}
        </div>
      </div>

      {/* Атрибуты оценки */}
      <div style={{ 
        display: 'flex', 
        gap: 'var(--space-4)',
        padding: 'var(--space-3)',
        background: 'var(--color-bg-soft)',
        borderRadius: 'var(--radius)',
      }}>
        <div style={{ flex: 1, textAlign: 'center' }}>
          <div style={{ 
            fontSize: 'var(--fs-xs)', 
            color: 'var(--color-text-muted)', 
            fontWeight: 500, 
            marginBottom: 'var(--space-1)' 
          }}>
            ЦЕННОСТЬ
          </div>
          <div style={{ 
            fontSize: 'var(--fs-xl)', 
            fontWeight: 600, 
            color: 'var(--color-text)' 
          }}>
            {formatAttribute(initiative.value)}
          </div>
        </div>
        
        <div style={{ flex: 1, textAlign: 'center' }}>
          <div style={{ 
            fontSize: 'var(--fs-xs)', 
            color: 'var(--color-text-muted)', 
            fontWeight: 500, 
            marginBottom: 'var(--space-1)' 
          }}>
            СКОРОСТЬ
          </div>
          <div style={{ 
            fontSize: 'var(--fs-xl)', 
            fontWeight: 600, 
            color: 'var(--color-text)' 
          }}>
            {formatAttribute(initiative.speed)}
          </div>
        </div>
        
        <div style={{ flex: 1, textAlign: 'center' }}>
          <div style={{ 
            fontSize: 'var(--fs-xs)', 
            color: 'var(--color-text-muted)', 
            fontWeight: 500, 
            marginBottom: 'var(--space-1)' 
          }}>
            СТОИМОСТЬ
          </div>
          <div style={{ 
            fontSize: 'var(--fs-xl)', 
            fontWeight: 600, 
            color: 'var(--color-text)' 
          }}>
            {formatAttribute(initiative.cost)}
          </div>
        </div>
      </div>

      {/* Описание */}
      {initiative.description && showFullDescription && (
        <div>
          <h3 style={{ 
            margin: 0, 
            marginBottom: 'var(--space-3)', 
            fontSize: 'var(--fs-lg)', 
            fontWeight: 600,
            color: 'var(--color-text)' 
          }}>
            Описание
          </h3>
          <div 
            dangerouslySetInnerHTML={{ __html: renderMarkdown(initiative.description) }}
            style={{
              lineHeight: 1.6,
              fontSize: 'var(--fs-md)',
              color: 'var(--color-text)',
              // Стили для Markdown элементов
              '& h1, & h2, & h3, & h4, & h5, & h6': {
                marginTop: 'var(--space-4)',
                marginBottom: 'var(--space-2)',
                color: 'var(--color-text)',
              },
              '& p': {
                marginBottom: 'var(--space-3)',
              },
              '& ul, & ol': {
                marginBottom: 'var(--space-3)',
                paddingLeft: 'var(--space-6)',
              },
              '& li': {
                marginBottom: 'var(--space-1)',
              },
              '& pre': {
                background: 'var(--color-bg-soft)',
                padding: 'var(--space-3)',
                borderRadius: 'var(--radius)',
                overflow: 'auto',
                fontSize: 'var(--fs-sm)',
                marginBottom: 'var(--space-3)',
              },
              '& code': {
                background: 'var(--color-bg-soft)',
                padding: '2px 4px',
                borderRadius: '4px',
                fontSize: 'var(--fs-sm)',
              },
              '& blockquote': {
                borderLeft: `4px solid var(--color-highlight)`,
                paddingLeft: 'var(--space-4)',
                marginLeft: 0,
                marginBottom: 'var(--space-3)',
                fontStyle: 'italic',
                color: 'var(--color-text-muted)',
              },
              '& a': {
                color: 'var(--color-highlight)',
                textDecoration: 'none',
              },
              '& a:hover': {
                textDecoration: 'underline',
              },
            } as React.CSSProperties}
          />
        </div>
      )}

      {/* Краткое описание для списков */}
      {initiative.description && !showFullDescription && (
        <div>
          <p style={{ 
            margin: 0, 
            color: 'var(--color-text-muted)', 
            fontSize: 'var(--fs-md)',
            lineHeight: 1.4,
            overflow: 'hidden',
            display: '-webkit-box',
            WebkitLineClamp: 3,
            WebkitBoxOrient: 'vertical',
          }}>
            {initiative.description.slice(0, 200)}
            {initiative.description.length > 200 && '...'}
          </p>
        </div>
      )}
      
      {/* Кнопка редактирования (если доступна) */}
      {onEdit && (
        <div style={{ 
          position: 'absolute', 
          top: 'var(--space-4)', 
          right: 'var(--space-4)' 
        }}>
          <button
            onClick={onEdit}
            style={{
              background: 'transparent',
              border: '1px solid var(--color-border)',
              borderRadius: 'var(--radius)',
              padding: 'var(--space-2)',
              cursor: 'pointer',
              color: 'var(--color-text-muted)',
              fontSize: 'var(--fs-sm)',
            }}
            title="Редактировать инициативу"
          >
            ✏️
          </button>
        </div>
      )}
    </Card>
  );
}