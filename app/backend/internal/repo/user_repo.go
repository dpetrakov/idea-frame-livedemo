package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository управляет операциями с пользователями в БД
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// CreateUser создает нового пользователя в БД
func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (login, display_name, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.pool.QueryRow(ctx, query,
		user.Login,
		user.DisplayName,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByLogin возвращает пользователя по логину
func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	user := &domain.User{}
	query := `
		SELECT id, login, display_name, password_hash, created_at, updated_at
		FROM users
		WHERE login = $1`

	err := r.pool.QueryRow(ctx, query, login).Scan(
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

	return user, nil
}

// GetUserByID возвращает пользователя по ID
func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	query := `
		SELECT id, login, display_name, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
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

	return user, nil
}

// GetAllUsers возвращает список всех пользователей (для назначения ответственных)
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]*domain.UserBrief, error) {
	query := `
		SELECT id, login, display_name
		FROM users
		ORDER BY display_name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*domain.UserBrief
	for rows.Next() {
		user := &domain.UserBrief{}
		err := rows.Scan(&user.ID, &user.Login, &user.DisplayName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}
