package http

import (
	"log/slog"
	"net/http"
	"time"
	
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/ideaframe/backend/internal/config"
	"github.com/ideaframe/backend/internal/http/handlers"
	"github.com/ideaframe/backend/internal/http/middleware"
	"github.com/ideaframe/backend/internal/repo"
	"github.com/ideaframe/backend/internal/service"
)

// NewRouter создаёт и настраивает HTTP роутер
func NewRouter(
	cfg *config.Config,
	logger *slog.Logger,
	db *repo.Database,
) http.Handler {
	r := chi.NewRouter()
	
	// Создаём репозитории
	initiativeRepo := repo.NewInitiativeRepository(db.Pool, logger)
	commentRepo := repo.NewCommentRepository(db.Pool, logger)
	
	// Создаём сервисы
	authService := service.NewAuthService(db, cfg.JWTSecret, cfg.JWTExpiration)
	userRepo := repo.NewUserRepository(db)
	initiativeService := service.NewInitiativeService(initiativeRepo, userRepo, logger)
	commentService := service.NewCommentService(commentRepo, initiativeRepo, logger)
	
	// Создаём handlers
	authHandler := handlers.NewAuthHandler(authService)
	healthHandler := handlers.NewHealthHandler(db)
	initiativeHandler := handlers.NewInitiativeHandlers(initiativeService)
	commentHandler := handlers.NewCommentHandlers(commentService)
	
	// Глобальные middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recoverer(logger))
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(middleware.CORS)
	
	// API версия 1
	r.Route("/api/v1", func(r chi.Router) {
		// Health check (публичный)
		r.Get("/health", healthHandler.Health)
		
		// Аутентификация (публичные)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})
		
		// Защищённые маршруты
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authService))
			
			// Пользователи
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", authHandler.GetCurrentUser)
				r.Get("/", authHandler.ListUsers)
			})
			
			// Инициативы
			r.Route("/initiatives", func(r chi.Router) {
				r.Get("/", initiativeHandler.ListInitiatives)       // GET /initiatives - список (подготовка к TK-005)
				r.Post("/", initiativeHandler.CreateInitiative)     // POST /initiatives - создание
				r.Get("/{id}", initiativeHandler.GetInitiative)     // GET /initiatives/{id} - детали
				r.Patch("/{id}", initiativeHandler.UpdateInitiative) // PATCH /initiatives/{id} - обновление (подготовка к TK-003, TK-006)
				// Комментарии к инициативе (TK-004)
				r.Get("/{id}/comments", commentHandler.ListComments)
				r.Post("/{id}/comments", commentHandler.CreateComment)
			})
		})
	})
	
	// Health check для /api/health (алиас для совместимости)
	r.Get("/api/health", healthHandler.Health)
	
	return r
}