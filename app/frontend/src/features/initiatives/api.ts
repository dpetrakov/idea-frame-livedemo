// API клиент для инициатив
// TK-002/FE: Функции для взаимодействия с backend endpoints

import { api } from '../../shared/lib/api-client';
import type { 
  Initiative, 
  InitiativeCreate, 
  InitiativeUpdate,
  InitiativesList 
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
 * Получение списка инициатив (подготовка для TK-005)
 * GET /v1/initiatives
 */
export async function getInitiativesList(params?: {
  filter?: string;
  limit?: number;
  offset?: number;
}): Promise<InitiativesList> {
  const searchParams = new URLSearchParams();
  
  if (params?.filter) {
    searchParams.append('filter', params.filter);
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