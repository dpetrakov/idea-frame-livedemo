package domain

import "github.com/google/uuid"

// ErrorResponse - единый формат ошибки согласно OpenAPI
type ErrorResponse struct {
	Code          string      `json:"code"`
	Message       string      `json:"message"`
	Details       interface{} `json:"details,omitempty"`
	CorrelationID string      `json:"correlationId,omitempty"`
}

// HealthResponse - ответ health check эндпоинта
type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Timestamp string `json:"timestamp"`
}

// NewErrorResponse создает новый ответ с ошибкой
func NewErrorResponse(code, message string, correlationID uuid.UUID) *ErrorResponse {
	return &ErrorResponse{
		Code:          code,
		Message:       message,
		CorrelationID: correlationID.String(),
	}
}

// NewErrorResponseWithDetails создает новый ответ с ошибкой и деталями
func NewErrorResponseWithDetails(code, message string, details interface{}, correlationID uuid.UUID) *ErrorResponse {
	return &ErrorResponse{
		Code:          code,
		Message:       message,
		Details:       details,
		CorrelationID: correlationID.String(),
	}
}