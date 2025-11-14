package api

type GetReviewResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}

type ReassignResponse struct {
	PullRequest PullRequestSchema `json:"pr"`
	ReplacedBy  string            `json:"replaced_by"`
}
