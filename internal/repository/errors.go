package repo

import "errors"

const (
	uniqueViolationCode = "23505"
)

var (
	ErrNotFound    = errors.New("resource not found")
	ErrTeamExists  = errors.New("team with this name already exists")
	ErrUserExists  = errors.New("user with this id already exists")
	ErrPRExists    = errors.New("PR id already exists")
	ErrPRMerged    = errors.New("cannot reassign on merged PR")
	ErrNotAssigned = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate = errors.New("no active replacement candidate in team")
)
