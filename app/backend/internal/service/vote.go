package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/ideaframe/backend/internal/domain"
	"github.com/ideaframe/backend/internal/repo"
)

// VoteService handles voting business logic
type VoteService struct {
	voteRepo       *repo.VoteRepository
	initiativeRepo *repo.InitiativeRepository
	logger         *slog.Logger
}

// NewVoteService creates new vote service
func NewVoteService(voteRepo *repo.VoteRepository, initiativeRepo *repo.InitiativeRepository, logger *slog.Logger) *VoteService {
	return &VoteService{
		voteRepo:       voteRepo,
		initiativeRepo: initiativeRepo,
		logger:         logger,
	}
}

// VoteForInitiative handles voting for an initiative
func (s *VoteService) VoteForInitiative(ctx context.Context, initiativeID uuid.UUID, req *domain.VoteRequest, userID uuid.UUID) (*domain.Initiative, error) {
	// Validate vote value
	if err := req.Validate(); err != nil {
		s.logger.Warn("invalid vote value", "error", err, "initiativeID", initiativeID, "userID", userID, "value", req.Value)
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if initiative exists and is not deleted
	_, err := s.initiativeRepo.GetByID(ctx, initiativeID)
	if err != nil {
		if err == domain.ErrInitiativeNotFound {
			s.logger.Warn("initiative not found for voting", "initiativeID", initiativeID, "userID", userID)
			return nil, domain.ErrInitiativeNotFound
		}
		s.logger.Error("failed to get initiative for voting", "error", err, "initiativeID", initiativeID, "userID", userID)
		return nil, fmt.Errorf("get initiative: %w", err)
	}

	// Get current user vote to log the change
	var prevValue int
	if len([]uuid.UUID{initiativeID}) > 0 {
		aggregates, err := s.voteRepo.GetVoteAggregates(ctx, []uuid.UUID{initiativeID}, userID)
		if err != nil {
			s.logger.Warn("failed to get current vote for logging", "error", err, "initiativeID", initiativeID, "userID", userID)
			prevValue = 0 // Default assumption
		} else if agg, exists := aggregates[initiativeID]; exists {
			prevValue = agg.CurrentUserVote
		}
	}

	// Apply the vote
	err = s.voteRepo.UpsertVote(ctx, initiativeID, userID, req.Value)
	if err != nil {
		s.logger.Error("failed to apply vote", "error", err, "initiativeID", initiativeID, "userID", userID, "value", req.Value)
		return nil, fmt.Errorf("apply vote: %w", err)
	}

	// Log the vote change
	s.logger.Info("vote_change",
		"initiativeId", initiativeID,
		"userId", userID,
		"prevValue", prevValue,
		"newValue", req.Value,
	)

	// Return updated initiative with new vote aggregates
	updatedInitiative, err := s.initiativeRepo.GetByIDWithVotes(ctx, initiativeID, userID, s.voteRepo)
	if err != nil {
		s.logger.Error("failed to get updated initiative after vote", "error", err, "initiativeID", initiativeID, "userID", userID)
		return nil, fmt.Errorf("get updated initiative: %w", err)
	}

	return updatedInitiative, nil
}
