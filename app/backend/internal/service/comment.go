package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/ideaframe/backend/internal/domain"
	"github.com/ideaframe/backend/internal/repo"
)

// CommentService бизнес-логика для комментариев
type CommentService struct {
	repo   *repo.CommentRepository
	inRepo *repo.InitiativeRepository
	logger *slog.Logger
}

func NewCommentService(repo *repo.CommentRepository, inRepo *repo.InitiativeRepository, logger *slog.Logger) *CommentService {
	return &CommentService{repo: repo, inRepo: inRepo, logger: logger}
}

// ListByInitiative возвращает список комментариев
func (s *CommentService) ListByInitiative(ctx context.Context, initiativeID uuid.UUID, limit, offset int) (*domain.CommentsList, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Проверим, что инициатива существует (404 если нет)
	if _, err := s.inRepo.GetByID(ctx, initiativeID); err != nil {
		if err == domain.ErrInitiativeNotFound {
			return nil, domain.ErrInitiativeNotFound
		}
		return nil, fmt.Errorf("check initiative: %w", err)
	}

	items, total, err := s.repo.ListByInitiative(ctx, initiativeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	return &domain.CommentsList{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

// Create создаёт новый комментарий
func (s *CommentService) Create(ctx context.Context, initiativeID, authorID uuid.UUID, req *domain.CommentCreate) (*domain.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, domain.ValidationError{Field: "text", Message: err.Error()}
	}

	// Проверим, что инициатива существует
	if _, err := s.inRepo.GetByID(ctx, initiativeID); err != nil {
		if err == domain.ErrInitiativeNotFound {
			return nil, domain.ErrInitiativeNotFound
		}
		return nil, fmt.Errorf("check initiative: %w", err)
	}

	c, err := s.repo.Create(ctx, initiativeID, authorID, req.Text)
	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	return c, nil
}
