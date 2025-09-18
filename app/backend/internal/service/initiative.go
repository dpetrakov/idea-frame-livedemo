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

// UpdateInitiative updates initiative attributes (prepared for TK-003 and TK-006)
func (s *InitiativeService) UpdateInitiative(ctx context.Context, id uuid.UUID, updates *domain.InitiativeUpdate, userID uuid.UUID) (*domain.Initiative, error) {
	// This method is prepared for TK-003 (attributes) and TK-006 (assignee) but not implemented in TK-002
	// For now, validate input and return the initiative as is
	if err := updates.Validate(); err != nil {
		s.logger.Warn("invalid update data", "error", err, "id", id, "userID", userID)
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Get existing initiative to check access
	initiative, err := s.GetInitiativeByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// For TK-002, just return the existing initiative
	// Full update logic will be implemented in TK-003 and TK-006
	s.logger.Info("initiative update requested (not implemented in TK-002)", "id", id, "userID", userID)
	return initiative, nil
}

// ListInitiatives returns initiatives list (prepared for TK-005)
func (s *InitiativeService) ListInitiatives(ctx context.Context, filter string, limit, offset int, userID uuid.UUID) ([]*domain.Initiative, int, error) {
	// This method is prepared for TK-005 but not implemented in TK-002
	// For now, return empty list
	s.logger.Info("initiative list requested (not implemented in TK-002)", "filter", filter, "userID", userID)
	return []*domain.Initiative{}, 0, nil
}