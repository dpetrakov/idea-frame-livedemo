package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ideaframe/backend/internal/domain"
	"github.com/ideaframe/backend/internal/http/middleware"
	"github.com/ideaframe/backend/internal/service"
	"github.com/ideaframe/backend/internal/telemetry"
)

// CommentHandlers HTTP-обработчики для комментариев
type CommentHandlers struct {
	service *service.CommentService
}

func NewCommentHandlers(service *service.CommentService) *CommentHandlers {
	return &CommentHandlers{service: service}
}

// ListComments GET /initiatives/{id}/comments
func (h *CommentHandlers) ListComments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := telemetry.LoggerFromContext(ctx)
	requestID := middleware.RequestIDFromContext(ctx)

	userID := middleware.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		middleware.RespondWithError(w, r, http.StatusUnauthorized, "Authorization required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	initiativeID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(w, r, http.StatusBadRequest, "Invalid initiative ID", "INVALID_ID")
		return
	}

	// Параметры пагинации
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, e := strconv.Atoi(l); e == nil { limit = v }
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, e := strconv.Atoi(o); e == nil { offset = v }
	}

	list, err := h.service.ListByInitiative(ctx, initiativeID, limit, offset)
	if err != nil {
		if errors.Is(err, domain.ErrInitiativeNotFound) {
			middleware.RespondWithError(w, r, http.StatusNotFound, "Initiative not found", "NOT_FOUND")
			return
		}
		logger.Error("failed to list comments", "error", err, "requestId", requestID)
		middleware.RespondWithError(w, r, http.StatusInternalServerError, "Failed to list comments", "INTERNAL_ERROR")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(list)
}

// CreateComment POST /initiatives/{id}/comments
func (h *CommentHandlers) CreateComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := telemetry.LoggerFromContext(ctx)
	requestID := middleware.RequestIDFromContext(ctx)

	userID := middleware.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		middleware.RespondWithError(w, r, http.StatusUnauthorized, "Authorization required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	initiativeID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.RespondWithError(w, r, http.StatusBadRequest, "Invalid initiative ID", "INVALID_ID")
		return
	}

	var req domain.CommentCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.RespondWithError(w, r, http.StatusBadRequest, "Invalid JSON", "INVALID_JSON")
		return
	}

	comment, err := h.service.Create(ctx, initiativeID, userID, &req)
	if err != nil {
		var vErr domain.ValidationError
		if errors.As(err, &vErr) {
			middleware.RespondWithError(w, r, http.StatusBadRequest, vErr.Error(), "VALIDATION_ERROR")
			return
		}
		if errors.Is(err, domain.ErrInitiativeNotFound) {
			middleware.RespondWithError(w, r, http.StatusNotFound, "Initiative not found", "NOT_FOUND")
			return
		}
		logger.Error("failed to create comment", "error", err, "requestId", requestID)
		middleware.RespondWithError(w, r, http.StatusInternalServerError, "Failed to create comment", "INTERNAL_ERROR")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}
