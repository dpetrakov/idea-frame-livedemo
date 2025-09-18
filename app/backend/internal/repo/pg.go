package repo

import (
	"context"
	"fmt"
	"time"
	
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database представляет подключение к БД
type Database struct {
	Pool *pgxpool.Pool
}

// Connect создаёт подключение к PostgreSQL
func Connect(ctx context.Context, dbURL string) (*Database, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}
	
	// Настройка пула соединений
	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}
	
	// Проверка соединения
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return &Database{Pool: pool}, nil
}

// MustConnect создаёт подключение или паникует
func MustConnect(ctx context.Context, dbURL string) *Database {
	db, err := Connect(ctx, dbURL)
	if err != nil {
		panic(err)
	}
	return db
}

// Close закрывает пул соединений
func (db *Database) Close() {
	db.Pool.Close()
}

// HealthCheck проверяет доступность БД
func (db *Database) HealthCheck(ctx context.Context) error {
	var result int
	err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	return err
}