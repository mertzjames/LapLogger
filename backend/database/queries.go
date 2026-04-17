package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/laplogger/laplogger/models"
)

// UpsertUser inserts a new user or updates their name/email if the google_id already exists.
// Returns the user record.
func (c *ControlDB) UpsertUser(googleID, email, name string) (*models.User, error) {
	id := uuid.New().String()
	now := time.Now().UTC()

	row := c.DB.QueryRow(`
		INSERT INTO users (id, google_id, email, name, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (google_id) DO UPDATE SET email = EXCLUDED.email, name = EXCLUDED.name
		RETURNING id, google_id, email, name, created_at
	`, id, googleID, email, name, now)

	var u models.User
	if err := row.Scan(&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}
	return &u, nil
}

// GetUserByID retrieves a user by their UUID.
func (c *ControlDB) GetUserByID(userID string) (*models.User, error) {
	var u models.User
	err := c.DB.QueryRow(
		`SELECT id, google_id, email, name, created_at FROM users WHERE id = $1`, userID,
	).Scan(&u.ID, &u.GoogleID, &u.Email, &u.Name, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}

// GetUserLeagues returns all leagues a user belongs to, along with their role.
func (c *ControlDB) GetUserLeagues(userID string) ([]models.League, error) {
	rows, err := c.DB.Query(`
		SELECT l.id, l.name, l.slug, l.db_name, l.created_at
		FROM leagues l
		JOIN league_memberships lm ON lm.league_id = l.id
		WHERE lm.user_id = $1
		ORDER BY l.name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get user leagues: %w", err)
	}
	defer rows.Close()

	var leagues []models.League
	for rows.Next() {
		var l models.League
		if err := rows.Scan(&l.ID, &l.Name, &l.Slug, &l.DBName, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan league: %w", err)
		}
		leagues = append(leagues, l)
	}
	return leagues, rows.Err()
}

// GetLeagueByID retrieves a single league by ID.
func (c *ControlDB) GetLeagueByID(leagueID string) (*models.League, error) {
	var l models.League
	err := c.DB.QueryRow(
		`SELECT id, name, slug, db_name, created_at FROM leagues WHERE id = $1`, leagueID,
	).Scan(&l.ID, &l.Name, &l.Slug, &l.DBName, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get league: %w", err)
	}
	return &l, nil
}

// CheckMembership returns true if the user is a member of the given league.
func (c *ControlDB) CheckMembership(userID, leagueID string) (bool, error) {
	var exists bool
	err := c.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM league_memberships WHERE user_id = $1 AND league_id = $2)`,
		userID, leagueID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check membership: %w", err)
	}
	return exists, nil
}
