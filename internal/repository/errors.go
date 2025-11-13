package repo

import "errors"

const (
	uniqueViolationCode = "23505"
)

var (
	ErrNotFound   = errors.New("resource not found")
	ErrTeamExists = errors.New("team with this name already exists")
	ErrUserExists = errors.New("user with this id already exists")
)
