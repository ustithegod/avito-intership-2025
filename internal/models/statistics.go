package models

type UserStatistics struct {
	UserID          string `db:"user_id"`
	Username        string `db:"username"`
	AssignmentCount int    `db:"assignment_count"`
}

type PrStatistics struct {
	PrCount   int `db:"pr_count"`
	OpenPrs   int `db:"open_pr_count"`
	MergedPrs int `db:"merged_pr_count"`
}
