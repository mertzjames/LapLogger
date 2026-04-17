package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

// TenantManager maintains a thread-safe pool of per-league database connections.
type TenantManager struct {
	mu       sync.RWMutex
	pools    map[string]*sql.DB
	user     string
	password string
	host     string
	port     string
}

// NewTenantManager creates a new tenant connection pool manager.
func NewTenantManager(user, password, host, port string) *TenantManager {
	return &TenantManager{
		pools:    make(map[string]*sql.DB),
		user:     user,
		password: password,
		host:     host,
		port:     port,
	}
}

// GetDB returns a database connection for the given tenant DB name.
// Connections are lazily initialized and cached.
func (tm *TenantManager) GetDB(dbName string) (*sql.DB, error) {
	// Fast path: read lock
	tm.mu.RLock()
	if db, ok := tm.pools[dbName]; ok {
		tm.mu.RUnlock()
		return db, nil
	}
	tm.mu.RUnlock()

	// Slow path: write lock, create connection
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Double-check after acquiring write lock
	if db, ok := tm.pools[dbName]; ok {
		return db, nil
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		tm.user, tm.password, tm.host, tm.port, dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open tenant db %s: %w", dbName, err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping tenant db %s: %w", dbName, err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)

	tm.pools[dbName] = db
	log.Printf("Connected to tenant database: %s", dbName)
	return db, nil
}

// Close closes all tenant database connections.
func (tm *TenantManager) Close() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for name, db := range tm.pools {
		if err := db.Close(); err != nil {
			log.Printf("Error closing tenant db %s: %v", name, err)
		}
	}
	tm.pools = make(map[string]*sql.DB)
}
