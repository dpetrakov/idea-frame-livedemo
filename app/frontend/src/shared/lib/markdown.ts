// Утилиты для работы с Markdown
// Безопасный рендеринг Markdown в HTML

import { marked } from 'marked';

// Настройка marked с безопасными параметрами
marked.setOptions({
  breaks: true,       // Переносы строк как <br>
  gfm: true,         // GitHub Flavored Markdown
});

/**
 * Рендерит Markdown в HTML с базовой санитизацией
 */
export function renderMarkdown(markdown: string): string {
  if (!markdown?.trim()) {
    return '';
  }
  
  try {
    const html = marked.parse(markdown) as string;
    // Базовая санитизация - удаление script тегов и onclick атрибутов
    return sanitizeHtml(html);
  } catch (error) {
    console.error('Error rendering markdown:', error);
    return escapeHtml(markdown);
  }
}

/**
 * Базовая санитизация HTML (для live-demo достаточно)
 */
function sanitizeHtml(html: string): string {
  return html
    .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '') // Удаляем script теги
    .replace(/on\w+="[^"]*"/gi, '') // Удаляем onclick и другие on* атрибуты
    .replace(/javascript:/gi, ''); // Удаляем javascript: ссылки
}

/**
 * Экранирует HTML символы
 */
function escapeHtml(text: string): string {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

/**
 * Преобразует HTML обратно в текст (для превью)
 */
export function htmlToText(html: string): string {
  const div = document.createElement('div');
  div.innerHTML = html;
  return div.textContent || div.innerText || '';
}

/**
 * Обрезает текст до указанной длины с учетом слов
 */
export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) {
    return text;
  }
  
  const truncated = text.substring(0, maxLength);
  const lastSpace = truncated.lastIndexOf(' ');
  
  if (lastSpace > 0) {
    return truncated.substring(0, lastSpace) + '…';
  }
  
  return truncated + '…';
}