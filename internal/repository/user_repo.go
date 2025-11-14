package repo

import (
	"context"
	"database/sql"
	"errors"

	"avito-intership-2025/internal/lib"
	"avito-intership-2025/internal/models"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	Save(ctx context.Context, user *models.User) (string, error)
	GetById(ctx context.Context, userID string) (*models.User, error)
	GetUsersInTeam(ctx context.Context, teamID int) ([]*models.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) error
}

type UserRepo struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewUserRepo(db *sqlx.DB, c *trmsqlx.CtxGetter) *UserRepo {
	return &UserRepo{
		db:     db,
		getter: c,
	}
}

func (r *UserRepo) Save(ctx context.Context, user *models.User) (string, error) {
	const op = "user_repo.Save"

	query := `
		INSERT INTO users (id, name, team_id, is_active, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			team_id = EXCLUDED.team_id,
			is_active = EXCLUDED.is_active
		RETURNING id;
	`

	var userID string
	err := r.getter.
		DefaultTrOrDB(ctx, r.db).
		QueryRowContext(ctx, query, user.ID, user.Name, user.TeamID, user.IsActive).Scan(&userID)
	if err != nil {
		return "", lib.Err(op, err)
	}

	return userID, nil
}

func (r *UserRepo) GetById(ctx context.Context, userID string) (*models.User, error) {
	const op = "user_repo.GetById"

	query := `
		SELECT id, name, team_id, is_active, created_at
		FROM users
		WHERE id = $1;
	`

	var user models.User
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &user, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, lib.Err(op, err)
	}

	return &user, nil
}

func (r *UserRepo) GetUsersInTeam(ctx context.Context, teamName string) ([]*models.User, error) {
	const op = "user_repo.GetUsersInTeam"

	query := `
		SELECT u.id, u.name, u.team_id, u.is_active, u.created_at
		FROM users u
		JOIN teams t ON u.team_id = t.id
		WHERE t.name = $1;
	`

	var users []*models.User
	err := r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &users, query, teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.User{}, nil
		}
		return nil, lib.Err(op, err)
	}

	return users, nil
}

func (r *UserRepo) GetActiveUsersIDInTeam(ctx context.Context, teamID int) ([]string, error) {
	const op = "user_repo.GetActiveUsersInTeam"

	query := `
		SELECT u.id
		FROM users u
		JOIN teams t ON u.team_id = t.id
		WHERE t.id = $1 AND u.is_active = TRUE;
	`

	var users []string
	err := r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &users, query, teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []string{}, nil
		}
		return nil, lib.Err(op, err)
	}

	return users, nil
}

func (r *UserRepo) SetIsActive(ctx context.Context, userID string, isActive bool) error {
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

/* заменил на Save, но не хочу удалять, т.к. мало ли понадобятся

func (r *UserRepo) Update(ctx context.Context, user *models.User) error {
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

func (r *UserRepo) Create(ctx context.Context, user *models.User) (string, error) {
	const op = "user_repo.Create"

	query := `
		INSERT INTO users (id, name, team_id, is_active, created_at)
		VALUES ($1, $2, $3, $4, now())
		RETURNING id;
	`

	var userID string
	err := r.getter.
		DefaultTrOrDB(ctx, r.db).
		QueryRowContext(ctx, query, user.ID, user.Name, user.TeamID, user.IsActive).Scan(&userID)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == uniqueViolationCode {
				return "", ErrUserExists
			}
		}
		return "", lib.Err(op, err)
	}

	return userID, nil
}
*/
