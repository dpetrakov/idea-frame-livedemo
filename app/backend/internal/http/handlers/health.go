package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	
	"github.com/ideaframe/backend/internal/repo"
)

// HealthResponse структура ответа health check
type HealthResponse struct {
	Status   string    `json:"status"`
	Database string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthHandler обработчик health check
type HealthHandler struct {
	db *repo.Database
}

// NewHealthHandler создаёт новый health handler
func NewHealthHandler(db *repo.Database) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// Health обработчик GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "ok",
		Database:  "connected",
		Timestamp: time.Now(),
	}
	
	// Проверяем подключение к БД
	if err := h.db.HealthCheck(r.Context()); err != nil {
		response.Status = "error"
		response.Database = "disconnected"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}