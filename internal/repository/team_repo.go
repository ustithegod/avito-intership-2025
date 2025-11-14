package repo

import (
	"context"
	"database/sql"
	"errors"

	"avito-intership-2025/internal/lib"
	"avito-intership-2025/internal/models"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type TeamRepository interface {
	Create(ctx context.Context, teamName string) (int, error)
	GetByTeamName(ctx context.Context, teamName string) (*models.Team, error)
}

type TeamRepo struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewTeamRepo(db *sqlx.DB, c *trmsqlx.CtxGetter) *TeamRepo {
	return &TeamRepo{
		db:     db,
		getter: c,
	}
}

func (r *TeamRepo) Create(ctx context.Context, teamName string) (int, error) {
	const op = "team_repo.Create"

	query := `
		INSERT INTO teams (name, created_at)
		VALUES ($1, now())
		RETURNING id;
	`

	var teamID int
	err := r.getter.DefaultTrOrDB(ctx, r.db).QueryRowContext(ctx, query, teamName).Scan(&teamID)
	if err != nil {
		pgErr := &pq.Error{}
		if errors.As(err, &pgErr) {
			if pgErr.Code == uniqueViolationCode {
				return 0, ErrTeamExists
			}
		}
		return 0, lib.Err(op, err)
	}

	return teamID, nil
}

func (r *TeamRepo) GetByTeamName(ctx context.Context, teamName string) (*models.Team, error) {
	const op = "team_repo.GetByTeamName"

	query := `
		SELECT id, name, created_at
		FROM teams
		WHERE name = $1;
	`

	var team models.Team
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &team, query, teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, lib.Err(op, err)
	}

	return &team, nil
}

func (r *TeamRepo) GetTeamNameByID(ctx context.Context, teamID int) (string, error) {
	const op = "team_repository.GetTeamNameByID"

	var teamName string
	query := `SELECT name FROM teams WHERE id = $1`
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &teamName, query, teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", lib.Err(op, err)
	}

	return teamName, nil
}
