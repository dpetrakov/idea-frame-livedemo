// API клиент для инициатив
// TK-002/FE: Функции для взаимодействия с backend endpoints

import { api } from '../../shared/lib/api-client';
import type { 
  Initiative, 
  InitiativeCreate, 
  InitiativeUpdate,
  InitiativesList,
  UserBrief,
  VoteRequest
} from './types';

/**
 * Создание новой инициативы
 * POST /v1/initiatives
 */
export async function createInitiative(data: InitiativeCreate): Promise<Initiative> {
  const response = await api<Initiative>('/v1/initiatives', {
    method: 'POST',
    body: JSON.stringify(data),
  });
  return response;
}

/**
 * Получение инициативы по ID
 * GET /v1/initiatives/{id}
 */
export async function getInitiativeById(id: string): Promise<Initiative> {
  const response = await api<Initiative>(`/v1/initiatives/${id}`, {
    method: 'GET',
  });
  return response;
}

/**
 * Обновление инициативы (подготовка для TK-003, TK-006)
 * PATCH /v1/initiatives/{id}
 */
export async function updateInitiative(id: string, data: InitiativeUpdate): Promise<Initiative> {
  const response = await api<Initiative>(`/v1/initiatives/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  });
  return response;
}

/**
 * Логическое удаление инициативы (только админ)
 * DELETE /v1/initiatives/{id}
 */
export async function deleteInitiative(id: string): Promise<void> {
  await api<void>(`/v1/initiatives/${id}`, {
    method: 'DELETE',
  });
}

/**
 * Получение списка инициатив (TK-005, TK-009)
 * GET /v1/initiatives
 */
export async function getInitiativesList(params?: {
  filter?: string;
  sort?: string;
  limit?: number;
  offset?: number;
}): Promise<InitiativesList> {
  const searchParams = new URLSearchParams();
  
  if (params?.filter) {
    searchParams.append('filter', params.filter);
  }
  if (params?.sort) {
    searchParams.append('sort', params.sort);
  }
  if (params?.limit !== undefined) {
    searchParams.append('limit', params.limit.toString());
  }
  if (params?.offset !== undefined) {
    searchParams.append('offset', params.offset.toString());
  }

  const queryString = searchParams.toString();
  const url = queryString ? `/v1/initiatives?${queryString}` : '/v1/initiatives';
  
  const response = await api<InitiativesList>(url, {
    method: 'GET',
  });
  return response;
}

/**
 * Получение списка пользователей (для выбора ответственного)
 * GET /v1/users
 */
export async function getUsers(): Promise<UserBrief[]> {
  return api<UserBrief[]>('/v1/users', { method: 'GET' });
}

/**
 * Комментарии к инициативе
 */
export async function getComments(initiativeId: string, params?: { limit?: number; offset?: number; }): Promise<import('./types').CommentsList> {
  const searchParams = new URLSearchParams();
  if (params?.limit !== undefined) searchParams.append('limit', String(params.limit));
  if (params?.offset !== undefined) searchParams.append('offset', String(params.offset));
  const qs = searchParams.toString();
  const url = qs ? `/v1/initiatives/${initiativeId}/comments?${qs}` : `/v1/initiatives/${initiativeId}/comments`;
  return api(url, { method: 'GET' });
}

export async function addComment(initiativeId: string, text: string): Promise<import('./types').Comment> {
  return api(`/v1/initiatives/${initiativeId}/comments`, {
    method: 'POST',
    body: JSON.stringify({ text }),
  });
}

/**
 * Голосование за инициативу
 * POST /v1/initiatives/{id}/vote
 */
export async function voteForInitiative(id: string, vote: VoteRequest): Promise<Initiative> {
  const response = await api<Initiative>(`/v1/initiatives/${id}/vote`, {
    method: 'POST',
    body: JSON.stringify(vote),
  });
  return response;
}

// Utility functions для валидации
export function validateInitiativeCreate(data: InitiativeCreate): string[] {
  const errors: string[] = [];
  
  if (!data.title?.trim()) {
    errors.push('Название обязательно');
  }
  
  if (data.title && data.title.length > 140) {
    errors.push('Название не должно превышать 140 символов');
  }
  
  if (data.description && data.description.length > 10000) {
    errors.push('Описание не должно превышать 10000 символов');
  }
  
  return errors;
}