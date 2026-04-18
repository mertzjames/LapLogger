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
	defer func() { _ = rows.Close() }()

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

// GetLeagueBySlug retrieves a single league by its URL slug.
func (c *ControlDB) GetLeagueBySlug(slug string) (*models.League, error) {
	var l models.League
	err := c.DB.QueryRow(
		`SELECT id, name, slug, db_name, created_at FROM leagues WHERE slug = $1`, slug,
	).Scan(&l.ID, &l.Name, &l.Slug, &l.DBName, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get league by slug: %w", err)
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

// CreateLeagueWithMembership atomically creates a league record with its db_name
// and adds the creator as league admin within a single transaction.
func (c *ControlDB) CreateLeagueWithMembership(name, slug, dbName, userID, role string) (*models.League, error) {
	tx, err := c.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	id := uuid.New().String()
	now := time.Now().UTC()

	var l models.League
	err = tx.QueryRow(`
		INSERT INTO leagues (id, name, slug, db_name, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, slug, db_name, created_at
	`, id, name, slug, dbName, now).Scan(&l.ID, &l.Name, &l.Slug, &l.DBName, &l.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create league: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO league_memberships (user_id, league_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
	`, userID, l.ID, role, now)
	if err != nil {
		return nil, fmt.Errorf("add membership: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return &l, nil
}

// DeleteLeague removes a league record and its memberships (used for rollback).
func (c *ControlDB) DeleteLeague(leagueID string) error {
	_, err := c.DB.Exec(`DELETE FROM leagues WHERE id = $1`, leagueID)
	if err != nil {
		return fmt.Errorf("delete league: %w", err)
	}
	return nil
}

// AddMembership creates a league membership for a user with the given role.
// Returns true if a new membership was created, false if it already existed.
func (c *ControlDB) AddMembership(userID, leagueID, role string) (bool, error) {
	result, err := c.DB.Exec(`
		INSERT INTO league_memberships (user_id, league_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`, userID, leagueID, role, time.Now().UTC())
	if err != nil {
		return false, fmt.Errorf("add membership: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}
	return rowsAffected > 0, nil
}
