package api

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	ErrInternalErr     = "INTERNAL_ERROR"
	ErrValidationErr   = "VALIDATION_ERROR"
	ErrBadRequest      = "BAD_REQUEST"
	ErrCodeNotFound    = "NOT_FOUND"
	ErrCodeTeamExists  = "TEAM_EXISTS"
	ErrCodePRExists    = "PR_EXISTS"
	ErrCodePRMerged    = "PR_MERGED"
	ErrCodeNotAssigned = "NOT_ASSIGNED"
	ErrCodeNoCandidate = "NO_CANDIDATE"
)

type TeamResponse struct {
	Team TeamSchema `json:"team"`
}

type UserResponse struct {
	User UserSchema `json:"user"`
}

type PrResponse struct {
	PullRequest PullRequestSchema `json:"pr"`
}

type GetReviewResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}

type ReassignResponse struct {
	PullRequest PullRequestSchema `json:"pr"`
	ReplacedBy  string            `json:"replaced_by"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type StatsResponse struct {
	Pr   PrStats     `json:"pr"`
	User []UserStats `json:"users"`
}

type UserStats struct {
	UserID          string `json:"user_id"`
	Username        string `json:"username"`
	AssignmentCount int    `json:"assignment_count"`
}

type PrStats struct {
	PrCount   int `json:"pr_count"`
	OpenPrs   int `json:"open_pr_count"`
	MergedPrs int `json:"merged_pr_count"`
}

func Error(code string, msg string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: msg,
		},
	}
}

func InternalError() ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    ErrInternalErr,
			Message: "internal server error",
		},
	}
}

func ValidationError(errs validator.ValidationErrors) ErrorResponse {
	var errMsgs []string
	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field '%s' is required", err.Field()))
		case "max":
			errMsgs = append(
				errMsgs,
				fmt.Sprintf("field '%s' must be no more than %s characters", err.Field(), err.Param()),
			)
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field '%s' is not valid", err.Field()))
		}
	}

	return ErrorResponse{
		Error: ErrorDetail{
			Code:    ErrValidationErr,
			Message: strings.Join(errMsgs, ", "),
		},
	}
}
