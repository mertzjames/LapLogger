package models

import "time"

// User represents a registered user in the control plane.
type User struct {
	ID        string    `json:"id" db:"id"`
	GoogleID  string    `json:"google_id" db:"google_id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// League represents a tenant (one league = one isolated database).
type League struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	DBName    string    `json:"-" db:"db_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// LeagueMembership links a user to a league with a role.
type LeagueMembership struct {
	UserID   string    `json:"user_id" db:"user_id"`
	LeagueID string    `json:"league_id" db:"league_id"`
	Role     string    `json:"role" db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}
