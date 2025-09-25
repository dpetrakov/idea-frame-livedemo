package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailCodeRepository struct {
	db *pgxpool.Pool
}

func NewEmailCodeRepository(db *pgxpool.Pool) *EmailCodeRepository {
	return &EmailCodeRepository{db: db}
}

// Create вставляет одноразовый код для e-mail
func (r *EmailCodeRepository) Create(ctx context.Context, email string, code string, ttl time.Duration, requestedIP string) error {
	const q = `
        INSERT INTO email_codes (email, code, expires_at, requested_ip)
        VALUES ($1, $2, NOW() + $3::interval, $4)
    `
	_, err := r.db.Exec(ctx, q, email, code, fmt.Sprintf("%f seconds", ttl.Seconds()), requestedIP)
	if err != nil {
		return fmt.Errorf("insert email code: %w", err)
	}
	return nil
}

// FindValid возвращает true, если существует актуальный неиспользованный код для пары email+code
func (r *EmailCodeRepository) FindValid(ctx context.Context, email string, code string) (bool, error) {
	const q = `
        SELECT 1
        FROM email_codes
        WHERE email = $1 AND code = $2 AND used_at IS NULL AND expires_at > NOW()
        LIMIT 1
    `
	var x int
	err := r.db.QueryRow(ctx, q, email, code).Scan(&x)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("find valid email code: %w", err)
	}
	return true, nil
}

// MarkUsed отмечает код как использованный
func (r *EmailCodeRepository) MarkUsed(ctx context.Context, email string, code string) error {
	const q = `
        UPDATE email_codes
        SET used_at = NOW()
        WHERE email = $1 AND code = $2 AND used_at IS NULL
    `
	ct, err := r.db.Exec(ctx, q, email, code)
	if err != nil {
		return fmt.Errorf("mark code used: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return nil
	}
	return nil
}

// LastRequestAt возвращает время последнего созданного кода для адреса
func (r *EmailCodeRepository) LastRequestAt(ctx context.Context, email string) (time.Time, error) {
	const q = `
        SELECT COALESCE(MAX(created_at), TO_TIMESTAMP(0))
        FROM email_codes
        WHERE email = $1
    `
	var t time.Time
	if err := r.db.QueryRow(ctx, q, email).Scan(&t); err != nil {
		return time.Time{}, fmt.Errorf("last request at: %w", err)
	}
	return t, nil
}
