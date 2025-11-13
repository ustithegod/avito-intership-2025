package main

import (
	"avito-intership-2025/internal/lib/config"
	"avito-intership-2025/internal/lib/sl"
	repo "avito-intership-2025/internal/repository"
	"avito-intership-2025/internal/service"
	"log/slog"
	"os"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
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

	trManager := manager.Must(trmsqlx.NewDefaultFactory(db))

	teamRepo := repo.NewTeamRepo(db, trmsqlx.DefaultCtxGetter)
	userRepo := repo.NewUserRepo(db, trmsqlx.DefaultCtxGetter)
	prRepo := repo.NewPullRequestRepo(db, trmsqlx.DefaultCtxGetter, trManager)

	teamService := service.NewTeamService(trManager, teamRepo, userRepo)
	userService := service.NewUserService(trManager, prRepo, userRepo, teamRepo)

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
