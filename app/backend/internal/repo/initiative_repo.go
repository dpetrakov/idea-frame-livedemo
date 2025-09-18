package repo

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ideaframe/backend/internal/domain"
)

// InitiativeRepository handles initiative data persistence
type InitiativeRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

// NewInitiativeRepository creates new initiative repository
func NewInitiativeRepository(db *pgxpool.Pool, logger *slog.Logger) *InitiativeRepository {
	return &InitiativeRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new initiative
func (r *InitiativeRepository) Create(ctx context.Context, initiative *domain.InitiativeCreate, authorID uuid.UUID) (*domain.Initiative, error) {
	const query = `
		INSERT INTO initiatives (title, description, author_id)
		VALUES ($1, $2, $3)
		RETURNING id, title, description, author_id, assignee_id, value, speed, cost, weight, created_at, updated_at
	`

	var result domain.Initiative
	var assigneeID sql.NullString

	err := r.db.QueryRow(ctx, query, initiative.Title, initiative.Description, authorID).Scan(
		&result.ID,
		&result.Title,
		&result.Description,
		&result.AuthorID,
		&assigneeID,
		&result.Value,
		&result.Speed,
		&result.Cost,
		&result.Weight,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("failed to create initiative", "error", err, "authorID", authorID)
		return nil, fmt.Errorf("create initiative: %w", err)
	}

	if assigneeID.Valid {
		assigneeUUID, _ := uuid.Parse(assigneeID.String)
		result.AssigneeID = &assigneeUUID
	}

	// Load author information
	author, err := r.loadUser(ctx, result.AuthorID)
	if err != nil {
		r.logger.Error("failed to load author", "error", err, "authorID", result.AuthorID)
		return nil, fmt.Errorf("load author: %w", err)
	}
	result.Author = *author

	// Load assignee if present
	if result.AssigneeID != nil {
		assignee, err := r.loadUser(ctx, *result.AssigneeID)
		if err != nil {
			r.logger.Warn("failed to load assignee", "error", err, "assigneeID", *result.AssigneeID)
		} else {
			result.Assignee = assignee
		}
	}

	// Set default values
	result.CommentsCount = 0 // No comments initially

	r.logger.Info("initiative created", "id", result.ID, "authorID", authorID, "title", result.Title)
	return &result, nil
}

// GetByID gets initiative by ID
func (r *InitiativeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Initiative, error) {
	const query = `
		SELECT i.id, i.title, i.description, i.author_id, i.assignee_id, 
		       i.value, i.speed, i.cost, i.weight, i.created_at, i.updated_at
		FROM initiatives i
		WHERE i.id = $1
	`

	var result domain.Initiative
	var assigneeID sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&result.ID,
		&result.Title,
		&result.Description,
		&result.AuthorID,
		&assigneeID,
		&result.Value,
		&result.Speed,
		&result.Cost,
		&result.Weight,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInitiativeNotFound
		}
		r.logger.Error("failed to get initiative", "error", err, "id", id)
		return nil, fmt.Errorf("get initiative: %w", err)
	}

	if assigneeID.Valid {
		assigneeUUID, _ := uuid.Parse(assigneeID.String)
		result.AssigneeID = &assigneeUUID
	}

	// Load author information
	author, err := r.loadUser(ctx, result.AuthorID)
	if err != nil {
		r.logger.Error("failed to load author", "error", err, "authorID", result.AuthorID)
		return nil, fmt.Errorf("load author: %w", err)
	}
	result.Author = *author

	// Load assignee if present
	if result.AssigneeID != nil {
		assignee, err := r.loadUser(ctx, *result.AssigneeID)
		if err != nil {
			r.logger.Warn("failed to load assignee", "error", err, "assigneeID", *result.AssigneeID)
		} else {
			result.Assignee = assignee
		}
	}

	// Set default values
	result.CommentsCount = 0 // Comments feature will be implemented in future tasks

	return &result, nil
}

// Update updates initiative fields (used for attributes and assignee updates in future tasks)
func (r *InitiativeRepository) Update(ctx context.Context, id uuid.UUID, updates *domain.InitiativeUpdate) (*domain.Initiative, error) {
	// This method is prepared for TK-003 and TK-006 but not implemented in TK-002
	// For now, return the initiative as is
	return r.GetByID(ctx, id)
}

// List gets initiatives with filtering and pagination (prepared for TK-005)
func (r *InitiativeRepository) List(ctx context.Context, filter string, limit, offset int) ([]*domain.Initiative, int, error) {
	// This method is prepared for TK-005 but not implemented in TK-002
	// For now, return empty list
	return []*domain.Initiative{}, 0, nil
}

// loadUser is a helper method to load user information
func (r *InitiativeRepository) loadUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	const query = `
		SELECT id, login, display_name, created_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Login,
		&user.DisplayName,
		&user.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("load user: %w", err)
	}

	return &user, nil
}