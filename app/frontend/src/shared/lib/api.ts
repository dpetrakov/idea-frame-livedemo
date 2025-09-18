// API клиент для взаимодействия с backend
export interface ApiError {
  code: string
  message: string
  details?: unknown
  correlationId?: string
}

export class ApiException extends Error {
  constructor(
    public status: number,
    public data: ApiError,
    message?: string
  ) {
    super(message || data.message || 'API Error')
    this.name = 'ApiException'
  }
}

// Базовый API клиент
export async function api<T>(
  path: string,
  init?: RequestInit & { token?: string }
): Promise<T> {
  const { token, ...requestInit } = init || {}
  
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(requestInit.headers as Record<string, string> || {}),
  }

  // Добавляем JWT токен если доступен
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  const baseUrl = import.meta.env.VITE_API_BASE || '/api'
  const url = `${baseUrl}/v1${path}`

  try {
    const response = await fetch(url, {
      ...requestInit,
      headers,
    })

    // Проверяем статус ответа
    if (!response.ok) {
      let errorData: ApiError
      try {
        errorData = await response.json()
      } catch {
        errorData = {
          code: 'UNKNOWN_ERROR',
          message: `HTTP ${response.status}: ${response.statusText}`,
        }
      }
      
      throw new ApiException(response.status, errorData)
    }

    // Парсим JSON ответ
    const data = await response.json()
    return data as T
  } catch (error) {
    if (error instanceof ApiException) {
      throw error
    }
    
    // Сетевые и другие ошибки
    throw new ApiException(0, {
      code: 'NETWORK_ERROR',
      message: error instanceof Error ? error.message : 'Ошибка сети',
    })
  }
}

// Удобные методы для разных HTTP методов
export const apiClient = {
  get: <T>(path: string, token?: string) => 
    api<T>(path, { method: 'GET', token }),
  
  post: <T>(path: string, data?: unknown, token?: string) =>
    api<T>(path, {
      method: 'POST',
      body: JSON.stringify(data),
      token,
    }),
  
  patch: <T>(path: string, data?: unknown, token?: string) =>
    api<T>(path, {
      method: 'PATCH',
      body: JSON.stringify(data),
      token,
    }),
  
  delete: <T>(path: string, token?: string) =>
    api<T>(path, { method: 'DELETE', token }),
}