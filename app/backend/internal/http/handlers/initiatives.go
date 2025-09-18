package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ideaframe/backend/internal/domain"
	"github.com/ideaframe/backend/internal/http/middleware"
	"github.com/ideaframe/backend/internal/service"
	"github.com/ideaframe/backend/internal/telemetry"
)

// InitiativeHandlers handles initiative HTTP endpoints
type InitiativeHandlers struct {
	initiativeService *service.InitiativeService
}

// NewInitiativeHandlers creates new initiative handlers
func NewInitiativeHandlers(initiativeService *service.InitiativeService) *InitiativeHandlers {
	return &InitiativeHandlers{
		initiativeService: initiativeService,
	}
}

// CreateInitiative handles POST /initiatives
func (h *InitiativeHandlers) CreateInitiative(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := telemetry.LoggerFromContext(ctx)
	requestID := middleware.RequestIDFromContext(ctx)

	// Get user ID from JWT context
	userID := middleware.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		logger.Warn("unauthorized access to create initiative", "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization required")
		return
	}

	// Parse request body
	var req domain.InitiativeCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid request body", "error", err, "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	// Create initiative
	initiative, err := h.initiativeService.CreateInitiative(ctx, &req, userID)
	if err != nil {
		if errors.Is(err, fmt.Errorf("validation error: %w", err)) {
			logger.Warn("validation error creating initiative", "error", err, "requestId", requestID)
			writeErrorResponse(w, requestID, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		
		logger.Error("failed to create initiative", "error", err, "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create initiative")
		return
	}

	logger.Info("initiative created", "id", initiative.ID, "authorID", userID, "requestId", requestID)

	// Return created initiative
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(initiative)
}

// GetInitiative handles GET /initiatives/{id}
func (h *InitiativeHandlers) GetInitiative(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := telemetry.LoggerFromContext(ctx)
	requestID := middleware.RequestIDFromContext(ctx)

	// Get user ID from JWT context
	userID := middleware.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		logger.Warn("unauthorized access to get initiative", "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization required")
		return
	}

	// Parse initiative ID from URL
	initiativeIDStr := chi.URLParam(r, "id")
	if initiativeIDStr == "" {
		logger.Warn("missing initiative ID in URL", "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusBadRequest, "INVALID_ID", "Initiative ID is required")
		return
	}

	initiativeID, err := uuid.Parse(initiativeIDStr)
	if err != nil {
		logger.Warn("invalid initiative ID format", "id", initiativeIDStr, "error", err, "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusBadRequest, "INVALID_ID", "Invalid initiative ID format")
		return
	}

	// Get initiative
	initiative, err := h.initiativeService.GetInitiativeByID(ctx, initiativeID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrInitiativeNotFound) {
			logger.Warn("initiative not found", "id", initiativeID, "userID", userID, "requestId", requestID)
			writeErrorResponse(w, requestID, http.StatusNotFound, "NOT_FOUND", "Initiative not found")
			return
		}
		
		logger.Error("failed to get initiative", "error", err, "id", initiativeID, "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve initiative")
		return
	}

	logger.Debug("initiative retrieved", "id", initiativeID, "userID", userID, "requestId", requestID)

	// Return initiative
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(initiative)
}

// UpdateInitiative handles PATCH /initiatives/{id} (TK-003, TK-006)
func (h *InitiativeHandlers) UpdateInitiative(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := telemetry.LoggerFromContext(ctx)
	requestID := middleware.RequestIDFromContext(ctx)

	// Get user ID from JWT context
	userID := middleware.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		logger.Warn("unauthorized access to update initiative", "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization required")
		return
	}

	// Parse initiative ID from URL
	initiativeIDStr := chi.URLParam(r, "id")
	if initiativeIDStr == "" {
		logger.Warn("missing initiative ID in URL", "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusBadRequest, "INVALID_ID", "Initiative ID is required")
		return
	}

	initiativeID, err := uuid.Parse(initiativeIDStr)
	if err != nil {
		logger.Warn("invalid initiative ID format", "id", initiativeIDStr, "error", err, "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusBadRequest, "INVALID_ID", "Invalid initiative ID format")
		return
	}

	// Parse request body
	var updates domain.InitiativeUpdate
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		logger.Warn("invalid request body", "error", err, "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON in request body")
		return
	}

	// Update initiative
	initiative, err := h.initiativeService.UpdateInitiative(ctx, initiativeID, &updates, userID)
	if err != nil {
		if errors.Is(err, domain.ErrInitiativeNotFound) {
			logger.Warn("initiative not found for update", "id", initiativeID, "userID", userID, "requestId", requestID)
			writeErrorResponse(w, requestID, http.StatusNotFound, "NOT_FOUND", "Initiative not found")
			return
		}
		var vErr domain.ValidationError
		if errors.As(err, &vErr) {
			logger.Warn("validation error updating initiative", "error", vErr, "id", initiativeID, "requestId", requestID)
			// details с указанием поля
			details := map[string]string{}
			if vErr.Field != "" {
				details[vErr.Field] = vErr.Message
			}
			middleware.RespondWithErrorDetails(w, r, http.StatusBadRequest, vErr.Error(), "VALIDATION_ERROR", details)
			return
		}
		
		logger.Error("failed to update initiative", "error", err, "id", initiativeID, "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update initiative")
		return
	}

	logger.Info("initiative updated", "id", initiativeID, "userID", userID, "requestId", requestID)

	// Return updated initiative
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(initiative)
}

// ListInitiatives handles GET /initiatives (TK-005)
func (h *InitiativeHandlers) ListInitiatives(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := telemetry.LoggerFromContext(ctx)
	requestID := middleware.RequestIDFromContext(ctx)

	// Get user ID from JWT context
	userID := middleware.UserIDFromContext(ctx)
	if userID == uuid.Nil {
		logger.Warn("unauthorized access to list initiatives", "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization required")
		return
	}

	// Parse query params
	q := r.URL.Query()
	filter := q.Get("filter")
	if filter == "" {
		filter = "all"
	}
	limit := 20
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	offset := 0
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	items, total, err := h.initiativeService.ListInitiatives(ctx, filter, limit, offset, userID)
	if err != nil {
		logger.Error("failed to list initiatives", "error", err, "requestId", requestID)
		writeErrorResponse(w, requestID, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list initiatives")
		return
	}

	resp := map[string]interface{}{
		"items":  items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	// Логируем основные параметры запроса списка
	logger.Info("initiatives_list",
		"filter", filter,
		"limit", limit,
		"offset", offset,
		"userID", userID,
		"count", len(items),
		"requestId", requestID,
	)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// writeErrorResponse writes standardized error response according to OpenAPI spec
func writeErrorResponse(w http.ResponseWriter, requestID string, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"code":          code,
		"message":       message,
		"correlationId": requestID,
	}

	json.NewEncoder(w).Encode(errorResponse)
}