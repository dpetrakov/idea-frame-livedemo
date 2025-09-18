package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/ideaframe/backend/internal/domain"
	"github.com/ideaframe/backend/internal/repo"
)

// InitiativeService handles initiative business logic
type InitiativeService struct {
	initiativeRepo *repo.InitiativeRepository
	logger         *slog.Logger
}

// NewInitiativeService creates new initiative service
func NewInitiativeService(initiativeRepo *repo.InitiativeRepository, logger *slog.Logger) *InitiativeService {
	return &InitiativeService{
		initiativeRepo: initiativeRepo,
		logger:         logger,
	}
}

// CreateInitiative creates a new initiative
func (s *InitiativeService) CreateInitiative(ctx context.Context, req *domain.InitiativeCreate, authorID uuid.UUID) (*domain.Initiative, error) {
	// Validate input
	if err := req.Validate(); err != nil {
		s.logger.Warn("invalid initiative data", "error", err, "authorID", authorID)
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Create initiative
	initiative, err := s.initiativeRepo.Create(ctx, req, authorID)
	if err != nil {
		s.logger.Error("failed to create initiative", "error", err, "authorID", authorID)
		return nil, fmt.Errorf("create initiative: %w", err)
	}

	s.logger.Info("initiative created successfully", 
		"id", initiative.ID, 
		"authorID", authorID, 
		"title", initiative.Title,
		"weight", initiative.Weight)

	return initiative, nil
}

// GetInitiativeByID retrieves initiative by ID
func (s *InitiativeService) GetInitiativeByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Initiative, error) {
	// Get initiative from repository
	initiative, err := s.initiativeRepo.GetByID(ctx, id)
	if err != nil {
		if err == domain.ErrInitiativeNotFound {
			s.logger.Warn("initiative not found", "id", id, "userID", userID)
			return nil, domain.ErrInitiativeNotFound
		}
		s.logger.Error("failed to get initiative", "error", err, "id", id, "userID", userID)
		return nil, fmt.Errorf("get initiative: %w", err)
	}

	// Check access permissions (for this demo, all users can see all initiatives)
	if !initiative.HasAccess(userID) {
		s.logger.Warn("access denied to initiative", "id", id, "userID", userID)
		return nil, domain.ErrInitiativeNotFound // Return not found instead of access denied for security
	}

	s.logger.Debug("initiative retrieved", "id", id, "userID", userID)
	return initiative, nil
}

// UpdateInitiative updates initiative attributes (TK-003) and will handle assignee in TK-006
func (s *InitiativeService) UpdateInitiative(ctx context.Context, id uuid.UUID, updates *domain.InitiativeUpdate, userID uuid.UUID) (*domain.Initiative, error) {
	// Валидация атрибутов (1-5 или null). Возвращаем доменную ошибку валидации
	if updates.Value != nil && (*updates.Value < 1 || *updates.Value > 5) {
		return nil, domain.ErrInvalidField("value", "must be between 1 and 5")
	}
	if updates.Speed != nil && (*updates.Speed < 1 || *updates.Speed > 5) {
		return nil, domain.ErrInvalidField("speed", "must be between 1 and 5")
	}
	if updates.Cost != nil && (*updates.Cost < 1 || *updates.Cost > 5) {
		return nil, domain.ErrInvalidField("cost", "must be between 1 and 5")
	}

	// Получаем текущую инициативу и проверяем доступ
	current, err := s.GetInitiativeByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Если нет изменений — no-op, возвращаем текущую
	if updates.Value == nil && updates.Speed == nil && updates.Cost == nil && updates.AssigneeID == nil {
		s.logger.Info("initiative update: no changes provided", "id", id, "userID", userID)
		return current, nil
	}

	// Обновляем через репозиторий (только переданные поля)
	updated, err := s.initiativeRepo.Update(ctx, id, updates)
	if err != nil {
		if err == domain.ErrInitiativeNotFound {
			return nil, domain.ErrInitiativeNotFound
		}
		return nil, fmt.Errorf("update initiative: %w", err)
	}

	// Логируем изменённые поля и новый вес
	changed := make([]string, 0, 3)
	if updates.Value != nil { changed = append(changed, "value") }
	if updates.Speed != nil { changed = append(changed, "speed") }
	if updates.Cost != nil { changed = append(changed, "cost") }
	s.logger.Info("initiative updated",
		"id", id,
		"userID", userID,
		"changed", changed,
		"weight", updated.Weight,
	)

	return updated, nil
}

// ListInitiatives returns initiatives list (prepared for TK-005)
func (s *InitiativeService) ListInitiatives(ctx context.Context, filter string, limit, offset int, userID uuid.UUID) ([]*domain.Initiative, int, error) {
	// This method is prepared for TK-005 but not implemented in TK-002
	// For now, return empty list
	s.logger.Info("initiative list requested (not implemented in TK-002)", "filter", filter, "userID", userID)
	return []*domain.Initiative{}, 0, nil
}