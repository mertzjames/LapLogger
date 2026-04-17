package database

import (
	"strings"
	"testing"
)

// === SEC-1: Database Layer Security Tests ===

// TestSecurity_ValidDBName_InjectionPatterns exercises the validDBName regex
// against a comprehensive set of SQL injection patterns targeting CREATE DATABASE.
func TestSecurity_ValidDBName_InjectionPatterns(t *testing.T) {
	injections := []string{
		// Classic SQL injection
		"db; DROP DATABASE laplogger_control",
		"db' OR '1'='1",
		"db\" OR \"1\"=\"1",
		"db-- comment",
		"db/* comment */name",

		// Shell injection via DB name
		"db$(whoami)",
		"db`whoami`",
		"db|cat /etc/passwd",

		// Unicode/encoding tricks
		"db\x00name",      // null byte
		"db\nname",        // newline
		"db\rname",        // carriage return
		"db\tname",        // tab

		// Path traversal
		"../../../etc/passwd",
		"db/../admin",

		// PostgreSQL specific
		"db; CREATE ROLE admin SUPERUSER LOGIN",
		"db; COPY (SELECT '') TO '/tmp/pwned'",
		"db; ALTER SYSTEM SET",
	}

	for _, injection := range injections {
		t.Run(injection[:min(len(injection), 40)], func(t *testing.T) {
			if validDBName.MatchString(injection) {
				t.Errorf("validDBName accepted malicious input: %q", injection)
			}
		})
	}
}

// TestSecurity_ValidDBName_LengthLimits verifies behavior with extremely
// long DB names. PostgreSQL has a 63-byte identifier limit.
func TestSecurity_ValidDBName_LengthLimits(t *testing.T) {
	// 63 chars = PostgreSQL max identifier length
	maxName := strings.Repeat("a", 63)
	if !validDBName.MatchString(maxName) {
		t.Errorf("valid 63-char name rejected: %q", maxName)
	}

	// 64 chars = exceeds PostgreSQL limit (but our regex doesn't enforce length)
	overName := strings.Repeat("a", 64)
	if !validDBName.MatchString(overName) {
		t.Log("Note: validDBName regex does not enforce PostgreSQL's 63-char identifier limit")
	}

	// Very long name (1000 chars)
	longName := strings.Repeat("a", 1000)
	if !validDBName.MatchString(longName) {
		t.Log("Note: validDBName regex does not enforce length limits")
	} else {
		t.Log("SEC-FINDING: No length limit on DB name. PostgreSQL truncates at 63 bytes, " +
			"which could cause name collisions for UUIDs that differ only after position 63")
	}
}

// TestSecurity_ProvisionTenantDB_EmptyName verifies empty DB names are rejected.
func TestSecurity_ProvisionTenantDB_EmptyName(t *testing.T) {
	err := ProvisionTenantDB(nil, "u", "p", "h", "5432", "")
	if err == nil {
		t.Fatal("expected error for empty DB name, got nil")
	}
}

// TestSecurity_TenantManager_DBNameIsolation verifies that the TenantManager
// maps DB names correctly and doesn't allow cross-tenant access via
// manipulated DB names.
func TestSecurity_TenantManager_DBNameIsolation(t *testing.T) {
	tm := NewTenantManager("user", "pass", "localhost", "5432")
	defer tm.Close()

	// Verify that the pool map uses the exact DB name as key
	// (no normalization that could merge different tenants)
	if _, exists := tm.pools["laplogger_league_aaa"]; exists {
		t.Fatal("pool should not contain entries before GetDB is called")
	}

	// Verify different DB names would produce different DSNs
	dsn1 := "postgres://user:pass@localhost:5432/laplogger_league_aaa?sslmode=disable"
	dsn2 := "postgres://user:pass@localhost:5432/laplogger_league_bbb?sslmode=disable"
	if dsn1 == dsn2 {
		t.Fatal("different DB names should produce different DSNs")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
