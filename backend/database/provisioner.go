package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"regexp"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

//go:embed migrations/tenant/*.sql
var tenantMigrations embed.FS

var validDBName = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// ProvisionTenantDB creates a new database for a league and runs tenant migrations.
func ProvisionTenantDB(controlDB *sql.DB, user, password, host, port, dbName string) error {
	if !validDBName.MatchString(dbName) {
		return fmt.Errorf("invalid database name: %s", dbName)
	}
	if len(dbName) > 63 {
		return fmt.Errorf("database name too long (%d chars, max 63): %s", len(dbName), dbName)
	}

	// Check if database already exists
	var exists bool
	err := controlDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check db exists: %w", err)
	}

	if !exists {
		// CREATE DATABASE cannot be parameterized, but we validated the name above
		_, err = controlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("create database: %w", err)
		}
		log.Printf("Created tenant database: %s", dbName)
	}

	// Connect to the new tenant DB and run migrations
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbName,
	)

	tenantDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open tenant db: %w", err)
	}
	defer func() { _ = tenantDB.Close() }()

	if err := runTenantMigrations(tenantDB); err != nil {
		return fmt.Errorf("tenant migrate: %w", err)
	}

	log.Printf("Tenant database %s migrated successfully", dbName)
	return nil
}

func runTenantMigrations(db *sql.DB) error {
	sourceDriver, err := iofs.New(tenantMigrations, "migrations/tenant")
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
