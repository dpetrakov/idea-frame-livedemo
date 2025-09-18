// Типы данных для инициатив согласно docs/openapi.yaml
// TK-002/FE: Типизация Initiative и связанных структур

export interface UserBrief {
  id: string;
  login: string;
  displayName: string;
}

export interface Initiative {
  id: string;
  title: string;
  description?: string | null;
  author: UserBrief;
  assignee?: UserBrief | null;
  value?: number | null; // 1-5
  speed?: number | null; // 1-5
  cost?: number | null;  // 1-5
  weight: number;        // Вычисляемый вес
  commentsCount: number; // Количество комментариев
  createdAt: string;     // ISO datetime
  updatedAt: string;     // ISO datetime
}

export interface InitiativeCreate {
  title: string;         // 1-140 символов
  description?: string | null; // до 10000 символов, опционально
}

export interface InitiativeUpdate {
  value?: number | null;      // 1-5 или null
  speed?: number | null;      // 1-5 или null  
  cost?: number | null;       // 1-5 или null
  assigneeId?: string | null; // UUID или null
}

export interface InitiativesList {
  items: Initiative[];
  total: number;
  limit: number;
  offset: number;
}

// Вспомогательные типы для UI
export type InitiativeFormData = {
  title: string;
  description: string;
};

export type InitiativeFormErrors = {
  title?: string;
  description?: string;
};

// Константы валидации
export const INITIATIVE_LIMITS = {
  TITLE_MIN: 1,
  TITLE_MAX: 140,
  DESCRIPTION_MAX: 10000,
  ATTRIBUTE_MIN: 1,
  ATTRIBUTE_MAX: 5,
} as const;

// Типы для состояний загрузки
export type LoadingState = 'idle' | 'loading' | 'success' | 'error';

export interface ApiError {
  code: string;
  message: string;
  correlationId?: string;
  details?: Record<string, any>;
}