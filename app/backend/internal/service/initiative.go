package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/ideaframe/backend/internal/config"
	"github.com/ideaframe/backend/internal/domain"
	"github.com/ideaframe/backend/internal/repo"
)

// InitiativeService handles initiative business logic
type InitiativeService struct {
	initiativeRepo *repo.InitiativeRepository
	userRepo       *repo.UserRepository
	voteRepo       *repo.VoteRepository
	logger         *slog.Logger
	cfg            *config.Config
}

// NewInitiativeService creates new initiative service
func NewInitiativeService(initiativeRepo *repo.InitiativeRepository, userRepo *repo.UserRepository, voteRepo *repo.VoteRepository, logger *slog.Logger, cfg *config.Config) *InitiativeService {
	return &InitiativeService{
		initiativeRepo: initiativeRepo,
		userRepo:       userRepo,
		voteRepo:       voteRepo,
		logger:         logger,
		cfg:            cfg,
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
	// Get initiative from repository with vote aggregates
	initiative, err := s.initiativeRepo.GetByIDWithVotes(ctx, id, userID, s.voteRepo)
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

// UpdateInitiative updates initiative attributes (TK-003) and assignee (TK-006)
func (s *InitiativeService) UpdateInitiative(ctx context.Context, id uuid.UUID, updates *domain.InitiativeUpdate, userID uuid.UUID) (*domain.Initiative, error) {
	// Валидация названия/описания
	if updates.Title != nil {
		trimmed := strings.TrimSpace(*updates.Title)
		if trimmed == "" {
			return nil, domain.ErrInvalidField("title", "must not be empty")
		}
		if utf8.RuneCountInString(trimmed) > 140 {
			return nil, domain.ErrInvalidField("title", "must not exceed 140 characters")
		}
		*updates.Title = trimmed
	}
	if updates.Description != nil {
		if utf8.RuneCountInString(*updates.Description) > 10000 {
			return nil, domain.ErrInvalidField("description", "must not exceed 10000 characters")
		}
	}

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

	// Проверяем права: изменение value/speed/cost/assigneeId только для админов
	if (updates.Value != nil || updates.Speed != nil || updates.Cost != nil || updates.AssigneeID.Present) && !s.isAdminByUserID(ctx, userID) {
		s.logger.Warn("forbidden update of admin fields", "initiativeId", id, "actorId", userID)
		return nil, domain.ErrForbidden
	}

	// Обработка assigneeId (валидация формата и существования пользователя)
	var newAssigneeID *uuid.UUID
	if updates.AssigneeID.Present {
		if updates.AssigneeID.Null {
			newAssigneeID = nil // снятие назначения
		} else {
			if !updates.AssigneeID.Valid {
				return nil, domain.ErrInvalidField("assigneeId", "invalid uuid")
			}
			// проверим существование пользователя
			u, err := s.userRepo.GetByID(ctx, updates.AssigneeID.Value)
			if err != nil {
				if err == domain.ErrNotFound || err == domain.ErrUserNotFound {
					return nil, domain.ErrInvalidField("assigneeId", "unknown user")
				}
				return nil, fmt.Errorf("check user exists: %w", err)
			}
			if u == nil {
				return nil, domain.ErrInvalidField("assigneeId", "unknown user")
			}
			// копируем для логов
			idCopy := updates.AssigneeID.Value
			newAssigneeID = &idCopy
		}
	}

	// Если нет изменений — no-op, возвращаем текущую
	noAssigneeChange := !updates.AssigneeID.Present
	noTitle := updates.Title == nil
	noDesc := updates.Description == nil
	if noTitle && noDesc && updates.Value == nil && updates.Speed == nil && updates.Cost == nil && noAssigneeChange {
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
	changed := make([]string, 0, 6)
	if updates.Title != nil {
		changed = append(changed, "title")
	}
	if updates.Description != nil {
		changed = append(changed, "description")
	}
	if updates.Value != nil {
		changed = append(changed, "value")
	}
	if updates.Speed != nil {
		changed = append(changed, "speed")
	}
	if updates.Cost != nil {
		changed = append(changed, "cost")
	}
	if updates.AssigneeID.Present {
		changed = append(changed, "assigneeId")
	}

	// Лог смены ответственного
	if updates.AssigneeID.Present {
		var oldID, newID string
		if current.AssigneeID != nil {
			oldID = current.AssigneeID.String()
		} else {
			oldID = "null"
		}
		if newAssigneeID != nil {
			newID = newAssigneeID.String()
		} else {
			newID = "null"
		}
		s.logger.Info("assignee changed",
			"initiativeId", id,
			"oldAssigneeId", oldID,
			"newAssigneeId", newID,
			"actorId", userID,
		)
	}

	s.logger.Info("initiative updated",
		"id", id,
		"userID", userID,
		"changed", changed,
		"weight", updated.Weight,
	)

	return updated, nil
}

// SoftDeleteInitiative логическое удаление инициативы (только админ)
func (s *InitiativeService) SoftDeleteInitiative(ctx context.Context, id uuid.UUID, actorID uuid.UUID) error {
	if !s.isAdminByUserID(ctx, actorID) {
		s.logger.Warn("forbidden delete initiative", "initiativeId", id, "actorId", actorID)
		return domain.ErrForbidden
	}

	deleted, err := s.initiativeRepo.SoftDelete(ctx, id)
	if err != nil {
		if err == domain.ErrInitiativeNotFound {
			return domain.ErrInitiativeNotFound
		}
		return fmt.Errorf("soft delete initiative: %w", err)
	}
	if deleted {
		s.logger.Info("initiative soft-deleted", "initiativeId", id, "actorId", actorID)
	} else {
		s.logger.Info("initiative already soft-deleted", "initiativeId", id, "actorId", actorID)
	}
	return nil
}

// isAdminByUserID возвращает true, если пользователь с данным ID является админом по email
func (s *InitiativeService) isAdminByUserID(ctx context.Context, userID uuid.UUID) bool {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || u == nil {
		return false
	}
	e := strings.ToLower(strings.TrimSpace(u.Email))
	for _, a := range s.cfg.AdminEmails {
		if e == a {
			return true
		}
	}
	return false
}

// ListInitiatives returns initiatives list (TK-005, TK-009)
func (s *InitiativeService) ListInitiatives(ctx context.Context, filter string, sort string, limit, offset int, userID uuid.UUID) ([]*domain.Initiative, int, error) {
	// Нормализация параметров
	switch filter {
	case "mineCreated", "assignedToMe", "all":
		// ok
	default:
		filter = "all"
	}
	switch sort {
	case "weight", "votes":
		// ok
	default:
		sort = "weight"
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := s.initiativeRepo.List(ctx, filter, sort, limit, offset, userID, s.voteRepo)
	if err != nil {
		return nil, 0, fmt.Errorf("list initiatives: %w", err)
	}
	return items, total, nil
}
