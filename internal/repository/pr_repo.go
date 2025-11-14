package repo

import (
	"avito-intership-2025/internal/lib"
	"avito-intership-2025/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PullRequestRepository interface {
	Create(ctx context.Context, pr *models.PullRequest) (string, error)
	GetById(ctx context.Context, prID string) (*models.PullRequest, error)
	GetByAuthor(ctx context.Context, authorID string) ([]*models.PullRequest, error)
	MarkAsMerged(ctx context.Context, prID string) error

	GetReviewers(ctx context.Context, prID string) ([]string, error)
	AssignReviewer(ctx context.Context, prID, userID string) error
	ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
}

type PullRequestRepo struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
	trm    *manager.Manager
}

func NewPullRequestRepo(db *sqlx.DB, c *trmsqlx.CtxGetter, trm *manager.Manager) *PullRequestRepo {
	return &PullRequestRepo{
		db:     db,
		getter: c,
		trm:    trm,
	}
}

func (r *PullRequestRepo) Create(ctx context.Context, pr *models.PullRequest) (string, error) {
	const op = "pull_request_repo.Create"

	query := `
        INSERT INTO pull_requests (id, title, author_id, status, created_at)
        VALUES ($1, $2, $3, $4, now())
        RETURNING id;
    `

	var prID string
	err := r.getter.DefaultTrOrDB(ctx, r.db).QueryRowContext(
		ctx,
		query,
		pr.ID,
		pr.Title,
		pr.AuthorId,
		pr.Status,
	).Scan(&prID)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == uniqueViolationCode {
				return "", ErrTeamExists
			}
		}
		return "", lib.Err(op, err)
	}

	return prID, nil
}

func (r *PullRequestRepo) GetById(ctx context.Context, prID string) (*models.PullRequest, error) {
	const op = "pull_request_repo.GetById"

	query := `
        SELECT id, title, author_id, status, created_at, merged_at
        FROM pull_requests
        WHERE id = $1
    `

	var pr models.PullRequest
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &pr, query, prID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, lib.Err(op, err)
	}

	return &pr, nil
}

func (r *PullRequestRepo) GetByAuthor(ctx context.Context, authorID string) ([]*models.PullRequest, error) {
	const op = "pull_request_repo.GetByAuthor"

	query := `
        SELECT id, title, author_id, status, created_at, merged_at
        FROM pull_requests
        WHERE author_id = $1
        ORDER BY created_at DESC
    `

	var prs []*models.PullRequest
	err := r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &prs, query, authorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.PullRequest{}, nil
		}
		return nil, lib.Err(op, err)
	}

	return prs, nil
}

func (r *PullRequestRepo) MarkAsMerged(ctx context.Context, prID string) error {
	const op = "pull_request_repo.MarkAsMerged"

	query := `
        UPDATE pull_requests
        SET status = 'MERGED', merged_at = now()
        WHERE id = $1
    `

	res, err := r.getter.DefaultTrOrDB(ctx, r.db).ExecContext(ctx, query, prID)
	if err != nil {
		return lib.Err(op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return lib.Err(op, err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PullRequestRepo) DeleteReviewer(ctx context.Context, prID, userID string) error {
	const op = "pull_request_repo.DeleteReviewer"

	res, err := r.getter.DefaultTrOrDB(ctx, r.db).ExecContext(
		ctx,
		`DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2`,
		prID, userID,
	)
	if err != nil {
		return lib.Err(op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return lib.Err(op, err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PullRequestRepo) GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequest, error) {
	const op = "pull_request_repo.GetUserReviews"

	query := `
		SELECT p.id, p.title, p.author_id, p.status, p.need_more_reviewers, p.created_at, p.merged_at
		FROM pull_requests p
		JOIN pr_reviewers prr ON prr.pull_request_id = p.id
		WHERE prr.user_id = $1
	`

	var pullRequests []*models.PullRequest
	err := r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &pullRequests, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, lib.Err(op, err)
	}

	return pullRequests, nil
}

func (r *PullRequestRepo) GetPrReviewers(ctx context.Context, prID string) ([]string, error) {
	const op = "pull_request_repo.GetReviewers"

	query := `
		SELECT user_id FROM pr_reviewers
		WHERE pull_request_id = $1;
	`

	var userIDs []string
	err := r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &userIDs, query, prID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []string{}, ErrNotFound
		}
		return nil, lib.Err(op, err)
	}

	return userIDs, nil
}

func (r *PullRequestRepo) AssignReviewer(ctx context.Context, prID, userID string) error {
	const op = "pull_request_repo.AssignReviewer"

	query := `
        INSERT INTO pr_reviewers (pull_request_id, user_id)
        VALUES ($1, $2)
    `

	_, err := r.getter.DefaultTrOrDB(ctx, r.db).ExecContext(ctx, query, prID, userID)
	if err != nil {
		return lib.Err(op, err)
	}

	return nil
}

func (r *PullRequestRepo) ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	const op = "pull_request_repo.ReassignReviewer"

	err := r.trm.Do(ctx, func(ctx context.Context) error {
		// Удаляем старого ревьюера
		res, err := r.db.ExecContext(ctx,
			`DELETE FROM pr_reviewers WHERE pull_request_id=$1 AND user_id=$2`,
			prID, oldUserID)
		if err != nil {
			return lib.Err(op, err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return ErrNotFound
		}

		if rowsAffected == 0 {
			return ErrNotFound
		}

		// Добавляем нового ревьюера
		_, err = r.db.ExecContext(ctx,
			`INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)`,
			prID, newUserID)
		if err != nil {
			// Проверяем на дубликат, если newUserID уже назначен
			if pgErr, ok := err.(*pq.Error); ok {
				if pgErr.Code == uniqueViolationCode {
					return fmt.Errorf("%s: new reviewer already assigned: %w", op, err)
				}
			}
			return lib.Err(op, err)
		}

		return nil
	})

	return err
}
