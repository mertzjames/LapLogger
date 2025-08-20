package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "laplogger.db")
	if err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	if err := seedData(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Swimmers table
		`CREATE TABLE IF NOT EXISTS swimmers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Strokes table
		`CREATE TABLE IF NOT EXISTS strokes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)`,

		// Meets table
		`CREATE TABLE IF NOT EXISTS meets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			location TEXT NOT NULL,
			meet_date DATE NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Events table (stroke + distance combinations)
		`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			stroke_id INTEGER NOT NULL,
			distance INTEGER NOT NULL,
			name TEXT NOT NULL,
			FOREIGN KEY (stroke_id) REFERENCES strokes(id),
			UNIQUE(stroke_id, distance)
		)`,

		// Meet Events table (which events are in which meets)
		`CREATE TABLE IF NOT EXISTS meet_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			meet_id INTEGER NOT NULL,
			event_id INTEGER NOT NULL,
			session TEXT NOT NULL,
			event_num INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (meet_id) REFERENCES meets(id),
			FOREIGN KEY (event_id) REFERENCES events(id),
			UNIQUE(meet_id, event_id, session)
		)`,

		// Swim Times table
		`CREATE TABLE IF NOT EXISTS swim_times (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			swimmer_id INTEGER NOT NULL,
			event_id INTEGER NOT NULL,
			meet_id INTEGER,
			time_ms INTEGER NOT NULL,
			notes TEXT,
			recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (swimmer_id) REFERENCES swimmers(id),
			FOREIGN KEY (event_id) REFERENCES events(id),
			FOREIGN KEY (meet_id) REFERENCES meets(id)
		)`,

		// Indexes for better performance
		`CREATE INDEX IF NOT EXISTS idx_swim_times_swimmer ON swim_times(swimmer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_swim_times_event ON swim_times(event_id)`,
		`CREATE INDEX IF NOT EXISTS idx_swim_times_meet ON swim_times(meet_id)`,
		`CREATE INDEX IF NOT EXISTS idx_meet_events_meet ON meet_events(meet_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func seedData(db *sql.DB) error {
	// Check if strokes already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM strokes").Scan(&count)
	if err != nil {
		return err
	}

	// Only seed if no strokes exist
	if count > 0 {
		return nil
	}

	// Seed strokes
	strokes := []string{
		"Freestyle",
		"Backstroke",
		"Breaststroke",
		"Butterfly",
		"Individual Medley",
	}

	for _, stroke := range strokes {
		_, err := db.Exec("INSERT INTO strokes (name) VALUES (?)", stroke)
		if err != nil {
			return err
		}
	}

	// Seed common events (stroke + distance combinations)
	events := []struct {
		strokeName string
		distance   int
	}{
		{"Freestyle", 50},
		{"Freestyle", 100},
		{"Freestyle", 200},
		{"Freestyle", 400},
		{"Freestyle", 800},
		{"Freestyle", 1500},
		{"Backstroke", 50},
		{"Backstroke", 100},
		{"Backstroke", 200},
		{"Breaststroke", 50},
		{"Breaststroke", 100},
		{"Breaststroke", 200},
		{"Butterfly", 50},
		{"Butterfly", 100},
		{"Butterfly", 200},
		{"Individual Medley", 200},
		{"Individual Medley", 400},
	}

	for _, event := range events {
		// Get stroke ID
		var strokeID int
		err := db.QueryRow("SELECT id FROM strokes WHERE name = ?", event.strokeName).Scan(&strokeID)
		if err != nil {
			return err
		}

		// Create event name
		eventName := fmt.Sprintf("%dm %s", event.distance, event.strokeName)

		// Insert event
		_, err = db.Exec("INSERT INTO events (stroke_id, distance, name) VALUES (?, ?, ?)",
			strokeID, event.distance, eventName)
		if err != nil {
			return err
		}
	}

	return nil
}
