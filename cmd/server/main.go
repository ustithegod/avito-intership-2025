package main

import (
	"avito-intership-2025/internal/http/handlers"
	prh "avito-intership-2025/internal/http/handlers/pr"
	teamh "avito-intership-2025/internal/http/handlers/team"
	userh "avito-intership-2025/internal/http/handlers/user"
	mw "avito-intership-2025/internal/http/middleware"
	"avito-intership-2025/internal/lib/config"
	"avito-intership-2025/internal/lib/sl"
	repo "avito-intership-2025/internal/repository"
	"avito-intership-2025/internal/service/pr"
	"avito-intership-2025/internal/service/team"
	"avito-intership-2025/internal/service/user"

	"log/slog"
	"net/http"
	"os"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("Starting PR Reviewer Assignment Service", slog.String("env", cfg.Env))

	dsn := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		slog.Error("failed to establish connection with database", sl.Err(err))
		os.Exit(1)
	}

	// initialization of go-transaction-manager
	trManager := manager.Must(trmsqlx.NewDefaultFactory(db))

	teamRepo := repo.NewTeamRepo(db, trmsqlx.DefaultCtxGetter)
	userRepo := repo.NewUserRepo(db, trmsqlx.DefaultCtxGetter)
	prRepo := repo.NewPullRequestRepo(db, trmsqlx.DefaultCtxGetter, trManager)

	teamService := team.NewTeamService(trManager, teamRepo, userRepo)
	userService := user.NewUserService(trManager, prRepo, userRepo, teamRepo)
	prService := pr.NewPullRequestService(trManager, prRepo, prRepo, userRepo)

	teamHandler := teamh.NewTeamHandler(log, teamService)
	userHandler := userh.NewUserHandler(log, userService)
	prHandler := prh.NewPrHandler(log, prService)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mw.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	log.Info("starting http server", slog.String("address", cfg.HTTPServer.Address))

	// public methods
	router.Get("/health", handlers.Healthcheck())
	router.Post("/team/add", teamHandler.Add)

	// user methods
	router.Group(func(r chi.Router) {
		r.Use(mw.AuthMiddleware)

		r.Get("/team/get", teamHandler.Get)
		r.Get("/users/getReview", userHandler.GetReview)
	})

	// admin methods
	router.Group(func(r chi.Router) {
		r.Use(mw.AuthMiddleware)
		r.Use(mw.AdminOnly)

		r.Post("/users/setIsActive", userHandler.SetIsActive)
		r.Post("/pullRequest/create", prHandler.Create)
		r.Post("/pullRequest/merge", prHandler.Merge)
		r.Post("/pullRequest/reassign", prHandler.Reassign)
	})

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start http server", sl.Err(err))
		os.Exit(1)
	}

	log.Error("http server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
