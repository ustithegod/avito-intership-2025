package models

import "time"

type PullRequest struct {
	ID                string     `db:"id"`
	Title             string     `db:"title"`
	AuthorId          string     `db:"author_id"`
	Status            string     `db:"status"`
	NeedMoreReviewers bool       `db:"need_more_reviewers"`
	CreatedAt         *time.Time `db:"created_at"`
}
