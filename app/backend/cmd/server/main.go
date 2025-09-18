package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/config"
	httpPkg "github.com/dpetrakov/idea-frame-livedemo/backend/internal/http"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/repo"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/security"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/service"
	"github.com/dpetrakov/idea-frame-livedemo/backend/internal/telemetry"
)

func main() {
	ctx := context.Background()

	// Загрузка конфигурации
	cfg := config.Load()
	
	// Инициализация логгера
	logger := telemetry.NewLogger(cfg)

	logger.Info("starting application",
		"port", cfg.Port,
		"env", cfg.Env,
	)

	// Подключение к БД и миграции
	pool := repo.MustConnect(ctx, cfg.DatabaseURL)
	defer pool.Close()

	// Инициализация сервисов
	jwtService := security.NewJWTService(cfg.JWTSecret)
	userRepo := repo.NewUserRepository(pool)
	authService := service.NewAuthService(userRepo, jwtService)

	// Создание HTTP роутера
	router := httpPkg.NewRouter(cfg, logger, pool, jwtService, authService)

	// HTTP сервер
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}

	// Запуск сервера в горутине
	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}