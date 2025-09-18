package repo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ideaframe/backend/internal/domain"
)

// CommentRepository отвечает за доступ к данным комментариев
type CommentRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewCommentRepository(db *pgxpool.Pool, logger *slog.Logger) *CommentRepository {
	return &CommentRepository{db: db, logger: logger}
}

// ListByInitiative возвращает список комментариев к инициативе с пагинацией по возрастанию created_at
func (r *CommentRepository) ListByInitiative(ctx context.Context, initiativeID uuid.UUID, limit, offset int) ([]domain.Comment, int, error) {
	const listQuery = `
		SELECT c.id, c.text, c.author_id, c.created_at
		FROM comments c
		WHERE c.initiative_id = $1
		ORDER BY c.created_at ASC
		LIMIT $2 OFFSET $3
	`

	const countQuery = `
		SELECT COUNT(*)
		FROM comments c
		WHERE c.initiative_id = $1
	`

	rows, err := r.db.Query(ctx, listQuery, initiativeID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Comment, 0, limit)
	for rows.Next() {
		var (
			c domain.Comment
			authorID uuid.UUID
		)
		if err := rows.Scan(&c.ID, &c.Text, &authorID, &c.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan comment: %w", err)
		}
		// Грузим компактные данные автора
		author, err := r.loadUserBrief(ctx, authorID)
		if err != nil {
			// Не срываем выдачу, но логируем
			r.logger.Warn("failed to load author for comment", "commentID", c.ID, "error", err)
			c.Author = domain.UserBrief{ID: authorID, Login: "unknown", DisplayName: "Unknown"}
		} else {
			c.Author = *author
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate comments: %w", err)
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, initiativeID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count comments: %w", err)
	}

	return items, total, nil
}

// Create создаёт комментарий к инициативе
func (r *CommentRepository) Create(ctx context.Context, initiativeID, authorID uuid.UUID, text string) (*domain.Comment, error) {
	const insertQuery = `
		INSERT INTO comments (initiative_id, author_id, text)
		VALUES ($1, $2, $3)
		RETURNING id, text, created_at
	`

	var c domain.Comment
	if err := r.db.QueryRow(ctx, insertQuery, initiativeID, authorID, text).Scan(&c.ID, &c.Text, &c.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert comment: %w", err)
	}

	// Подтягиваем автора
	author, err := r.loadUserBrief(ctx, authorID)
	if err != nil {
		r.logger.Warn("failed to load author after insert", "commentID", c.ID, "error", err)
		c.Author = domain.UserBrief{ID: authorID, Login: "unknown", DisplayName: "Unknown"}
	} else {
		c.Author = *author
	}

	return &c, nil
}

// loadUserBrief грузит краткую информацию о пользователе
func (r *CommentRepository) loadUserBrief(ctx context.Context, userID uuid.UUID) (*domain.UserBrief, error) {
	const q = `
		SELECT id, login, display_name
		FROM users
		WHERE id = $1
	`
	var u domain.UserBrief
	if err := r.db.QueryRow(ctx, q, userID).Scan(&u.ID, &u.Login, &u.DisplayName); err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("load user brief: %w", err)
	}
	return &u, nil
}
