package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/ideaframe/backend/internal/config"
	httpserver "github.com/ideaframe/backend/internal/http"
	"github.com/ideaframe/backend/internal/repo"
	"github.com/ideaframe/backend/internal/telemetry"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Инициализация логгера
	logger := telemetry.NewLogger(cfg.LogLevel, cfg.Environment)
	
	// Контекст для инициализации
	ctx := context.Background()
	
	// Подключение к БД
	logger.Info("Connecting to database...")
	db, err := repo.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("Database connected")
	
	// Создание роутера
	router := httpserver.NewRouter(cfg, logger, db)
	
	// Создание HTTP сервера
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}
	
	// Запуск сервера в отдельной горутине
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.Port, "env", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()
	
	// Ожидание сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	
	// Graceful shutdown
	logger.Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown server gracefully", "error", err)
		os.Exit(1)
	}
	
	logger.Info("Server stopped")
}