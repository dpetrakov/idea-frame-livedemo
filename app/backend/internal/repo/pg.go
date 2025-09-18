package repo

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// MustConnect создает пул соединений с PostgreSQL и запускает миграции
func MustConnect(ctx context.Context, databaseURL string) *pgxpool.Pool {
	// Подключение к БД
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}

	// Проверка подключения
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Запуск миграций
	if err := runMigrations(databaseURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database connected and migrations applied successfully")
	return pool
}

// runMigrations применяет миграции БД
func runMigrations(databaseURL string) error {
	// Парсим строку подключения
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("could not parse database URL: %w", err)
	}

	// Подключение для миграций
	db := stdlib.OpenDB(*config)
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create database driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not apply migrations: %w", err)
	}

	log.Println("Database migrations applied successfully")
	return nil
}