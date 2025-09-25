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
	result.UpVotes = 0
	result.DownVotes = 0
	result.VoteScore = 0
	result.CurrentUserVote = 0

	r.logger.Info("initiative created", "id", result.ID, "authorID", authorID, "title", result.Title)
	return &result, nil
}

// GetByID gets initiative by ID
func (r *InitiativeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Initiative, error) {
	const query = `
        SELECT i.id, i.title, i.description, i.author_id, i.assignee_id, 
               i.value, i.speed, i.cost, i.weight, i.created_at, i.updated_at
        FROM initiatives i
        WHERE i.id = $1 AND i.is_deleted = false
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
	result.UpVotes = 0
	result.DownVotes = 0
	result.VoteScore = 0
	result.CurrentUserVote = 0

	return &result, nil
}

// GetByIDWithVotes gets initiative by ID with vote aggregates
func (r *InitiativeRepository) GetByIDWithVotes(ctx context.Context, id uuid.UUID, userID uuid.UUID, voteRepo *VoteRepository) (*domain.Initiative, error) {
	initiative, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load vote aggregates
	aggregates, err := voteRepo.GetVoteAggregates(ctx, []uuid.UUID{id}, userID)
	if err != nil {
		r.logger.Warn("failed to load vote aggregates", "error", err, "initiativeID", id)
		// Continue without vote data - vote fields will remain 0
	} else if agg, exists := aggregates[id]; exists {
		initiative.UpVotes = agg.UpVotes
		initiative.DownVotes = agg.DownVotes
		initiative.VoteScore = agg.VoteScore
		initiative.CurrentUserVote = agg.CurrentUserVote
	}

	return initiative, nil
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

	query := fmt.Sprintf("UPDATE initiatives SET %s WHERE id = $%d AND is_deleted = false", strings.Join(setParts, ", "), idx)
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
func (r *InitiativeRepository) List(ctx context.Context, filter string, sort string, limit, offset int, userID uuid.UUID, voteRepo *VoteRepository) ([]*domain.Initiative, int, error) {
	where := "WHERE i.is_deleted = false"
	args := []any{}
	argIdx := 1

	switch filter {
	case "mineCreated":
		where = fmt.Sprintf("%s AND i.author_id = $%d", where, argIdx)
		args = append(args, userID)
		argIdx++
	case "assignedToMe":
		where = fmt.Sprintf("%s AND i.assignee_id = $%d", where, argIdx)
		args = append(args, userID)
		argIdx++
	default:
		// all — только фильтр по is_deleted
	}

	countQuery := "SELECT COUNT(*) FROM initiatives i " + where

	// Define sort order
	var orderBy string
	switch sort {
	case "votes":
		orderBy = "ORDER BY vote_score DESC, i.created_at DESC"
	default: // "weight"
		orderBy = "ORDER BY i.weight DESC, i.created_at DESC"
	}

	listQuery := `
		SELECT 
		  i.id, i.title, i.description, i.author_id, i.assignee_id,
		  i.value, i.speed, i.cost, i.weight,
		  a.id, a.login, a.display_name, a.created_at,
		  assign.id, assign.login, assign.display_name,
		  COALESCE(c.cnt, 0) AS comments_count,
		  COALESCE(v.up_votes, 0) AS up_votes,
		  COALESCE(v.down_votes, 0) AS down_votes,
		  COALESCE(v.up_votes, 0) - COALESCE(v.down_votes, 0) AS vote_score,
		  i.created_at, i.updated_at
		FROM initiatives i
		JOIN users a ON a.id = i.author_id
		LEFT JOIN users assign ON assign.id = i.assignee_id
		LEFT JOIN LATERAL (
		  SELECT COUNT(*)::int AS cnt FROM comments cm WHERE cm.initiative_id = i.id
		) c ON true
		LEFT JOIN LATERAL (
		  SELECT 
		    COUNT(*) FILTER (WHERE value = 1) AS up_votes,
		    COUNT(*) FILTER (WHERE value = -1) AS down_votes
		  FROM initiative_votes iv WHERE iv.initiative_id = i.id
		) v ON true
	` + " " + where + " " + orderBy + " LIMIT %d OFFSET %d"

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
			authorID      uuid.UUID
			authorLogin   string
			authorDisplay string
			authorCreated time.Time
			// assignee brief (nullable)
			assignID      sql.NullString
			assignLogin   sql.NullString
			assignDisplay sql.NullString
			commentsCount int
			upVotes       int
			downVotes     int
			voteScore     int
		)

		if err := rows.Scan(
			&init.ID, &init.Title, &init.Description, &init.AuthorID, &assigneeID,
			&init.Value, &init.Speed, &init.Cost, &init.Weight,
			&authorID, &authorLogin, &authorDisplay, &authorCreated,
			&assignID, &assignLogin, &assignDisplay,
			&commentsCount, &upVotes, &downVotes, &voteScore,
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
		init.UpVotes = upVotes
		init.DownVotes = downVotes
		init.VoteScore = voteScore
		// CurrentUserVote will be loaded separately below

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

	// Load current user votes for all initiatives
	if len(items) > 0 && voteRepo != nil {
		ids := make([]uuid.UUID, len(items))
		for i, item := range items {
			ids[i] = item.ID
		}

		aggregates, err := voteRepo.GetVoteAggregates(ctx, ids, userID)
		if err != nil {
			r.logger.Warn("failed to load current user votes for list", "error", err, "userID", userID)
		} else {
			for _, item := range items {
				if agg, exists := aggregates[item.ID]; exists {
					item.CurrentUserVote = agg.CurrentUserVote
				}
			}
		}
	}

	return items, total, nil
}

// SoftDelete помечает инициативу как удалённую. Возвращает (deleted=true), если была изменена запись.
// Если инициатива уже удалена, возвращает (deleted=false, nil). Если не существует — ErrInitiativeNotFound.
func (r *InitiativeRepository) SoftDelete(ctx context.Context, id uuid.UUID) (bool, error) {
	const delQuery = `
        UPDATE initiatives
        SET is_deleted = true, updated_at = NOW()
        WHERE id = $1 AND is_deleted = false
    `
	ct, err := r.db.Exec(ctx, delQuery, id)
	if err != nil {
		return false, fmt.Errorf("soft delete initiative: %w", err)
	}
	if ct.RowsAffected() > 0 {
		return true, nil
	}
	// Проверим, существует ли запись
	const existsQuery = `SELECT 1 FROM initiatives WHERE id = $1`
	var dummy int
	if err := r.db.QueryRow(ctx, existsQuery, id).Scan(&dummy); err != nil {
		if err == pgx.ErrNoRows {
			return false, domain.ErrInitiativeNotFound
		}
		return false, fmt.Errorf("check initiative exists: %w", err)
	}
	// Существует, но уже удалена
	return false, nil
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
