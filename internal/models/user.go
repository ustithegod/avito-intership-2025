package models

import (
	"time"
)

type User struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	TeamID    int        `db:"team_id"`
	IsActive  bool       `db:"is_active"`
	CreatedAt *time.Time `db:"created_at"`
}
