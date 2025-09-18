package repo

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

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

// Update updates initiative fields (used for attributes and assignee updates)
func (r *InitiativeRepository) Update(ctx context.Context, id uuid.UUID, updates *domain.InitiativeUpdate) (*domain.Initiative, error) {
	setParts := make([]string, 0, 4)
	args := make([]interface{}, 0, 5)
	idx := 1

if updates.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", idx))
		args = append(args, strings.TrimSpace(*updates.Title))
		idx++
	}
	if updates.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", idx))
		args = append(args, *updates.Description)
		idx++
	}
	if updates.Value != nil {
		setParts = append(setParts, fmt.Sprintf("value = $%d", idx))
		args = append(args, *updates.Value)
		idx++
	}
	if updates.Speed != nil {
		setParts = append(setParts, fmt.Sprintf("speed = $%d", idx))
		args = append(args, *updates.Speed)
		idx++
	}
	if updates.Cost != nil {
		setParts = append(setParts, fmt.Sprintf("cost = $%d", idx))
		args = append(args, *updates.Cost)
		idx++
	}
	// TK-006: поддержка изменения assignee_id
	if updates.AssigneeID.Present {
		if updates.AssigneeID.Null {
			setParts = append(setParts, "assignee_id = NULL")
		} else if updates.AssigneeID.Valid {
			setParts = append(setParts, fmt.Sprintf("assignee_id = $%d", idx))
			args = append(args, updates.AssigneeID.Value)
			idx++
		}
	}

	if len(setParts) == 0 {
		// Нет изменений — возвращаем текущую запись
		return r.GetByID(ctx, id)
	}

	query := fmt.Sprintf("UPDATE initiatives SET %s WHERE id = $%d", strings.Join(setParts, ", "), idx)
	args = append(args, id)

	ct, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update initiative", "error", err, "id", id)
		return nil, fmt.Errorf("update initiative: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return nil, domain.ErrInitiativeNotFound
	}

	// Возвращаем обновлённую запись с полями автора/ответственного
	return r.GetByID(ctx, id)
}

// List gets initiatives with filtering and pagination (TK-005)
func (r *InitiativeRepository) List(ctx context.Context, filter string, limit, offset int, userID uuid.UUID) ([]*domain.Initiative, int, error) {
	where := ""
	args := []any{}
	argIdx := 1

	switch filter {
	case "mineCreated":
		where = fmt.Sprintf("WHERE i.author_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	case "assignedToMe":
		where = fmt.Sprintf("WHERE i.assignee_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	default:
		// all — без WHERE
	}

	countQuery := "SELECT COUNT(*) FROM initiatives i " + where

	listQuery := `
		SELECT 
		  i.id, i.title, i.description, i.author_id, i.assignee_id,
		  i.value, i.speed, i.cost, i.weight,
		  a.id, a.login, a.display_name, a.created_at,
		  assign.id, assign.login, assign.display_name,
		  COALESCE(c.cnt, 0) AS comments_count,
		  i.created_at, i.updated_at
		FROM initiatives i
		JOIN users a ON a.id = i.author_id
		LEFT JOIN users assign ON assign.id = i.assignee_id
		LEFT JOIN LATERAL (
		  SELECT COUNT(*)::int AS cnt FROM comments cm WHERE cm.initiative_id = i.id
		) c ON true
	` + " " + where + " ORDER BY i.weight DESC, i.created_at DESC LIMIT %d OFFSET %d"

	// Выполняем запрос списка
	finalListQuery := fmt.Sprintf(listQuery, limit, offset)

	rows, err := r.db.Query(ctx, finalListQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list initiatives: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.Initiative, 0, limit)
	for rows.Next() {
		var (
			init       domain.Initiative
			assigneeID sql.NullString
			// author fields (with created_at)
			authorID uuid.UUID
			authorLogin string
			authorDisplay string
			authorCreated time.Time
			// assignee brief (nullable)
			assignID sql.NullString
			assignLogin sql.NullString
			assignDisplay sql.NullString
			commentsCount int
		)

		if err := rows.Scan(
			&init.ID, &init.Title, &init.Description, &init.AuthorID, &assigneeID,
			&init.Value, &init.Speed, &init.Cost, &init.Weight,
			&authorID, &authorLogin, &authorDisplay, &authorCreated,
			&assignID, &assignLogin, &assignDisplay,
			&commentsCount,
			&init.CreatedAt, &init.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan initiative: %w", err)
		}

		// Author
		init.Author = domain.User{
			ID:          authorID,
			Login:       authorLogin,
			DisplayName: authorDisplay,
			CreatedAt:   authorCreated,
		}

		// Assignee
		if assigneeID.Valid && assignID.Valid {
			assigneeUUID, _ := uuid.Parse(assigneeID.String)
			init.AssigneeID = &assigneeUUID
			init.Assignee = &domain.User{
				ID:          assigneeUUID,
				Login:       assignLogin.String,
				DisplayName: assignDisplay.String,
				// CreatedAt оставим нулевым, он не используется на клиенте
			}
		}

		init.CommentsCount = commentsCount

		items = append(items, &init)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate initiatives: %w", err)
	}

	// Считаем total
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count initiatives: %w", err)
	}

	return items, total, nil
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