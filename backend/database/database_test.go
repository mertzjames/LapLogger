package database

import (
	"testing"
)

func TestValidDBName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid simple", "laplogger_league_abc123", true},
		{"valid underscores", "laplogger_league_550e8400_e29b_41d4", true},
		{"valid alphanumeric", "testdb123", true},
		{"empty string", "", false},
		{"contains dash", "laplogger-league-abc", false},
		{"contains space", "laplogger league", false},
		{"contains semicolon", "db_name;DROP TABLE users", false},
		{"contains dot", "public.users", false},
		{"contains slash", "db/name", false},
		{"contains single quote", "db'name", false},
		{"sql injection attempt", "test; DROP DATABASE laplogger_control;--", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validDBName.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("validDBName.MatchString(%q) = %v; want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewTenantManager(t *testing.T) {
	tm := NewTenantManager("user", "pass", "localhost", "5432")
	if tm == nil {
		t.Fatal("NewTenantManager returned nil")
	}
	if tm.user != "user" {
		t.Errorf("user = %q; want %q", tm.user, "user")
	}
	if tm.password != "pass" {
		t.Errorf("password = %q; want %q", tm.password, "pass")
	}
	if tm.host != "localhost" {
		t.Errorf("host = %q; want %q", tm.host, "localhost")
	}
	if tm.port != "5432" {
		t.Errorf("port = %q; want %q", tm.port, "5432")
	}
	if tm.pools == nil {
		t.Error("pools map is nil")
	}
	if len(tm.pools) != 0 {
		t.Errorf("pools length = %d; want 0", len(tm.pools))
	}
}

func TestTenantManager_Close_Empty(t *testing.T) {
	tm := NewTenantManager("user", "pass", "localhost", "5432")
	tm.Close()
	if len(tm.pools) != 0 {
		t.Errorf("pools length after close = %d; want 0", len(tm.pools))
	}
}

func TestProvisionTenantDB_InvalidName(t *testing.T) {
	err := ProvisionTenantDB(nil, "user", "pass", "host", "5432", "invalid-name-with-dashes")
	if err == nil {
		t.Fatal("expected error for invalid DB name, got nil")
	}
	expected := "invalid database name: invalid-name-with-dashes"
	if err.Error() != expected {
		t.Errorf("error = %q; want %q", err.Error(), expected)
	}
}

func TestProvisionTenantDB_SQLInjectionName(t *testing.T) {
	err := ProvisionTenantDB(nil, "u", "p", "h", "5432", "db; DROP TABLE users;--")
	if err == nil {
		t.Fatal("expected error for SQL injection attempt, got nil")
	}
}
