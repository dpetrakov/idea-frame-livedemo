package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/config"
	httpserver "github.com/dpetrakov/idea-frame-livedemo/backend/internal/http"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/repo"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/service"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/telemetry"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	// Загружаем .env файл если он есть (для локальной разработки)
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Warning: failed to load .env file: %v\n", err)
		}
	}

	// Загрузка конфигурации
	cfg := config.Load()

	// Настройка логгера
	logger := telemetry.NewLogger(cfg.Env)
	logger.Info("starting server", "port", cfg.Port, "env", cfg.Env)

	// Выполнение миграций БД
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("database migrations completed successfully")

	// Подключение к базе данных
	pool := repo.MustConnect(ctx, cfg.DatabaseURL)
	defer pool.Close()
	logger.Info("database connected successfully")

	// Создание сервисов
	userRepo := repo.NewUserRepository(pool)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)

	// Создание HTTP сервера
	router := httpserver.NewRouter(cfg, logger, authService, pool)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Запуск сервера в отдельной горутине
	go func() {
		logger.Info("http server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}

// runMigrations выполняет миграции базы данных с помощью golang-migrate
func runMigrations(databaseURL string) error {
	// Путь к директории с миграциями (будет скопирован в Docker образ)
	migrationsPath := "/app/db/migrations"

	cmd := exec.Command("migrate", "-database", databaseURL, "-path", migrationsPath, "up")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Проверяем если ошибка о том, что нет изменений - это не критично
		if string(output) == "no change\n" ||
			len(output) == 0 {
			return nil
		}
		return fmt.Errorf("migrate command failed: %w, output: %s", err, string(output))
	}

	fmt.Printf("Migration output: %s\n", string(output))
	return nil
}
