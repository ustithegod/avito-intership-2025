package repo

import (
	"avito-intership-2025/internal/lib"
	"avito-intership-2025/internal/models"
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (string, error) {
	const op = "user_repo.Create"

	query := `
		INSERT INTO users (id, name, team_id, is_active, created_at)
		VALUES ($1, $2, $3, $4, now())
		RETURNING id;
	`

	var userID string
	err := r.db.QueryRowContext(ctx, query, user.ID, user.Name, user.TeamID, user.IsActive).Scan(&userID)
	if err != nil {
		return "", lib.Err(op, err)
	}

	return userID, nil
}

func (r *UserRepository) GetById(ctx context.Context, userID string) (*models.User, error) {
	const op = "user_repo.GetById"

	query := `
		SELECT id, name, team_id, is_active, created_at
		FROM users
		WHERE id = $1;
	`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, lib.Err(op, err)
	}

	return &user, nil
}

func (r *UserRepository) GetUsersInTeam(ctx context.Context, teamID int) ([]*models.User, error) {
	const op = "user_repo.GetUsersInTeam"

	query := `
		SELECT id, name, team_id, is_active, created_at
		FROM users
		WHERE team_id = $1;
	`

	var users []*models.User
	err := r.db.SelectContext(ctx, &users, query, teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.User{}, nil
		}
		return nil, lib.Err(op, err)
	}

	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	const op = "user_repo.Update"

	query := `
		UPDATE users
		SET name = $1, team_id = $2, is_active = $3
		WHERE id = $4;
	`

	res, err := r.db.ExecContext(ctx, query, user.Name, user.TeamID, user.IsActive, user.ID)
	if err != nil {
		return lib.Err(op, err)
	}

	// проверяем поменялся ли пользователь
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return lib.Err(op, err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	const op = "user_repo.SetIsActive"

	query := `UPDATE users SET is_active = $1 WHERE id = $2`

	res, err := r.db.ExecContext(ctx, query, isActive, userID)
	if err != nil {
		return lib.Err(op, err)
	}

	// проверяем сработал ли запрос на какую-либо строку
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return lib.Err(op, err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
