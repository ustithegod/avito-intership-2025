package repo

import (
	"avito-intership-2025/internal/lib"
	"avito-intership-2025/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PullRequestRepository struct {
	db *sqlx.DB
}

func NewPullRequestRepository(db *sqlx.DB) *PullRequestRepository {
	return &PullRequestRepository{
		db: db,
	}
}

func (r *PullRequestRepository) Create(ctx context.Context, pr *models.PullRequest) (string, error) {
	const op = "pull_request_repo.Create"

	query := `
        INSERT INTO pull_requests (id, title, author_id, status, need_more_reviewers, created_at)
        VALUES ($1, $2, $3, $4, $5, now())
        RETURNING id;
    `

	var prID string
	err := r.db.QueryRowContext(
		ctx,
		query,
		pr.ID,
		pr.Title,
		pr.AuthorId,
		pr.Status,
		pr.NeedMoreReviewers,
	).Scan(&prID)

	if err != nil {
		return "", lib.Err(op, err)
	}

	return prID, nil
}

func (r *PullRequestRepository) GetById(ctx context.Context, prID string) (*models.PullRequest, error) {
	const op = "pull_request_repo.GetById"

	query := `
        SELECT id, title, author_id, status, need_more_reviewers, created_at
        FROM pull_requests
        WHERE id = $1
    `

	var pr models.PullRequest
	err := r.db.GetContext(ctx, &pr, query, prID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, lib.Err(op, err)
	}

	return &pr, nil
}

func (r *PullRequestRepository) GetByAuthor(ctx context.Context, authorID string) ([]*models.PullRequest, error) {
	const op = "pull_request_repo.GetByAuthor"

	query := `
        SELECT id, title, author_id, status, need_more_reviewers, created_at
        FROM pull_requests
        WHERE author_id = $1
        ORDER BY created_at DESC
    `

	var prs []*models.PullRequest
	err := r.db.SelectContext(ctx, &prs, query, authorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.PullRequest{}, nil
		}
		return nil, lib.Err(op, err)
	}

	return prs, nil
}

func (r *PullRequestRepository) MarkAsMerged(ctx context.Context, prID string) error {
	const op = "pull_request_repo.MarkAsMerged"

	query := `
        UPDATE pull_requests
        SET status = 'MERGED'
        WHERE id = $1
    `

	res, err := r.db.ExecContext(ctx, query, prID)
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

func (r *PullRequestRepository) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	const op = "pull_request_repo.GetReviewers"

	query := `
		SELECT user_id FROM pr_reviewers
		WHERE pull_request_id = $1;
	`

	var userIDs []string
	err := r.db.SelectContext(ctx, &userIDs, query, prID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, lib.Err(op, err)
	}

	return userIDs, nil
}

func (r *PullRequestRepository) AssignReviewer(ctx context.Context, prID, userID string) error {
	const op = "pull_request_repo.AssignReviewer"

	query := `
        INSERT INTO pr_reviewers (pull_request_id, user_id)
        VALUES ($1, $2)
    `

	_, err := r.db.ExecContext(ctx, query, prID, userID)
	if err != nil {
		return lib.Err(op, err)
	}

	return nil
}

func (r *PullRequestRepository) ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	const op = "pull_request_repo.ReassignReviewer"

	// Начинаем транзакцию, т.к. операция должна быть атомарной
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return lib.Err(op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Удаляем старого ревьюера
	res, err := tx.ExecContext(ctx,
		`DELETE FROM pr_reviewers WHERE pull_request_id=$1 AND user_id=$2`,
		prID, oldUserID)
	if err != nil {
		return lib.Err(op, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return lib.Err(op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: old reviewer not assigned: %w", op, ErrNotFound)
	}

	// Добавляем нового ревьюера
	_, err = tx.ExecContext(ctx,
		`INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)`,
		prID, newUserID)
	if err != nil {
		// проверяем на дубликат, если newUserID уже назначен
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == uniqueViolationCode {
				return fmt.Errorf("%s: new reviewer already assigned: %w", op, err)
			}
		}
		return lib.Err(op, err)
	}

	return nil
}
