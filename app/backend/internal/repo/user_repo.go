package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// CreateUser создает нового пользователя
func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (login, display_name, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, login, display_name, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, user.Login, user.DisplayName, user.PasswordHash)

	var createdUser domain.User
	createdUser.PasswordHash = user.PasswordHash

	err := row.Scan(
		&createdUser.ID,
		&createdUser.Login,
		&createdUser.DisplayName,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Код ошибки уникальности в PostgreSQL
			if pgErr.Code == "23505" { // unique_violation
				return nil, domain.ErrUserAlreadyExists
			}
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &createdUser, nil
}

// GetUserByLogin получает пользователя по логину
func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	query := `
		SELECT id, login, display_name, password_hash, created_at, updated_at
		FROM users
		WHERE login = $1`

	row := r.pool.QueryRow(ctx, query, login)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Login,
		&user.DisplayName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	return &user, nil
}

// GetUserByID получает пользователя по ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, login, display_name, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, userID)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Login,
		&user.DisplayName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}