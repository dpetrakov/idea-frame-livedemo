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
  upVotes: number;       // Количество голосов "вверх"
  downVotes: number;     // Количество голосов "вниз"
  voteScore: number;     // upVotes - downVotes
  currentUserVote: number; // 1, -1, или 0 (нет голоса)
  commentsCount: number; // Количество комментариев
  createdAt: string;     // ISO datetime
  updatedAt: string;     // ISO datetime
}

export interface InitiativeCreate {
  title: string;         // 1-140 символов
  description?: string | null; // до 10000 символов, опционально
}

export interface InitiativeUpdate {
  title?: string;             // 1-140 символов
  description?: string | null; // до 10000 символов, можно пустую строку
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

// Комментарии
export interface Comment {
  id: string;
  text: string;
  author: UserBrief;
  createdAt: string;
}

export interface CommentsList {
  items: Comment[];
  total: number;
  limit: number;
  offset: number;
}

// Голосование
export interface VoteRequest {
  value: -1 | 0 | 1; // -1 (down), 0 (remove), 1 (up)
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