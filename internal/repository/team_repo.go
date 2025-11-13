package repo

import (
	"avito-intership-2025/internal/lib"
	"avito-intership-2025/internal/models"
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{
		db: db,
	}
}

func (r *TeamRepository) Create(ctx context.Context, teamName string) (int, error) {
	const op = "team_repo.Create"

	query := `
		INSERT INTO teams (name, created_at)
		VALUES ($1, now())
		RETURNING id;
	`

	var teamID int
	err := r.db.QueryRowContext(ctx, query, teamName).Scan(&teamID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == uniqueViolationCode {
				return 0, ErrTeamExists
			}
		}
		return 0, lib.Err(op, err)
	}

	return teamID, nil
}

func (r *TeamRepository) GetByTeamName(ctx context.Context, teamName string) (*models.Team, error) {
	const op = "team_repo.GetByTeamName"

	query := `
		SELECT id, name, created_at
		FROM teams
		WHERE name = $1;
	`

	var team models.Team
	err := r.db.GetContext(ctx, &team, query, teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, lib.Err(op, err)
	}

	return &team, nil
}
