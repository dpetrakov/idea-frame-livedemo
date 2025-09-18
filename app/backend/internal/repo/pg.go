package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MustConnect создает пул соединений с PostgreSQL или завершает работу при ошибке
func MustConnect(ctx context.Context, databaseURL string) *pgxpool.Pool {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse database URL: %v", err))
	}

	// Настройки пула соединений
	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		panic(fmt.Sprintf("failed to create connection pool: %v", err))
	}

	// Проверка подключения
	if err := pool.Ping(ctx); err != nil {
		panic(fmt.Sprintf("failed to ping database: %v", err))
	}

	return pool
}

// HealthCheck проверяет состояние подключения к БД
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
