package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ideaframe/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// UserRepository репозиторий для работы с пользователями
type UserRepository struct {
	db *Database
}

// NewUserRepository создаёт новый репозиторий пользователей
func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{db: db}
}

// Create создаёт нового пользователя
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, login, display_name, password_hash, email, email_verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	user.ID = uuid.New()

	_, err := r.db.Pool.Exec(ctx, query,
		user.ID,
		user.Login,
		user.DisplayName,
		user.PasswordHash,
		user.Email,
		user.EmailVerifiedAt,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		// Проверка на дублирование логина/почты
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "idx_users_login" {
					return domain.ErrConflict
				}
				if pgErr.ConstraintName == "idx_users_email" {
					return domain.ErrConflict
				}
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByLogin получает пользователя по логину
func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	query := `
		SELECT id, login, display_name, password_hash, email, email_verified_at, created_at, updated_at
		FROM users
		WHERE login = $1
	`

	user := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Email,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	return user, nil
}

// GetByID получает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, login, display_name, password_hash, email, email_verified_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Login,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Email,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail получает пользователя по e-mail
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
        SELECT id, login, display_name, password_hash, email, email_verified_at, created_at, updated_at
        FROM users
        WHERE email = $1
    `

	user := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Login,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Email,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// List получает список всех пользователей (для выбора ответственных)
func (r *UserRepository) List(ctx context.Context) ([]domain.UserBrief, error) {
	query := `
		SELECT id, login, display_name
		FROM users
		ORDER BY display_name, login
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []domain.UserBrief
	for rows.Next() {
		var user domain.UserBrief
		if err := rows.Scan(&user.ID, &user.Login, &user.DisplayName); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}
