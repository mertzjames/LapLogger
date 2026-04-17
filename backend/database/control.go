package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrations/control/*.sql
var controlMigrations embed.FS

// ControlDB holds the connection to the control plane database.
type ControlDB struct {
	DB *sql.DB
}

// NewControlDB connects to the control plane database and runs migrations.
func NewControlDB(user, password, host, port, dbname string) (*ControlDB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("control db open: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("control db ping: %w", err)
	}

	if err := runControlMigrations(db); err != nil {
		return nil, fmt.Errorf("control db migrate: %w", err)
	}

	log.Println("Control DB connected and migrated")
	return &ControlDB{DB: db}, nil
}

func runControlMigrations(db *sql.DB) error {
	sourceDriver, err := iofs.New(controlMigrations, "migrations/control")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migration db driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

// DSN returns a connection string for a given database name on the same server.
func (c *ControlDB) DSN(user, password, host, port, dbname string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname,
	)
}
