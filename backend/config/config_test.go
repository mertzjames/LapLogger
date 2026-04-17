package config

import (
	"os"
	"testing"
)

func TestGetEnv_Fallback(t *testing.T) {
	key := "LAPLOGGER_TEST_NONEXISTENT_KEY"
	_ = os.Unsetenv(key)

	got := getEnv(key, "default_val")
	if got != "default_val" {
		t.Errorf("getEnv(%q, %q) = %q; want %q", key, "default_val", got, "default_val")
	}
}

func TestGetEnv_SetValue(t *testing.T) {
	key := "LAPLOGGER_TEST_CONFIG_KEY"
	t.Setenv(key, "custom_val")

	got := getEnv(key, "default_val")
	if got != "custom_val" {
		t.Errorf("getEnv(%q, %q) = %q; want %q", key, "default_val", got, "custom_val")
	}
}

func TestLoad_Defaults(t *testing.T) {
	for _, k := range []string{
		"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_HOST",
		"POSTGRES_PORT", "POSTGRES_DB", "JWT_SECRET", "BACKEND_PORT",
	} {
		t.Setenv(k, "")
		_ = os.Unsetenv(k)
	}

	cfg := Load()

	if cfg.PostgresUser != "laplogger" {
		t.Errorf("PostgresUser = %q; want %q", cfg.PostgresUser, "laplogger")
	}
	if cfg.PostgresPassword != "laplogger_dev" {
		t.Errorf("PostgresPassword = %q; want %q", cfg.PostgresPassword, "laplogger_dev")
	}
	if cfg.PostgresHost != "localhost" {
		t.Errorf("PostgresHost = %q; want %q", cfg.PostgresHost, "localhost")
	}
	if cfg.PostgresPort != "5432" {
		t.Errorf("PostgresPort = %q; want %q", cfg.PostgresPort, "5432")
	}
	if cfg.PostgresDB != "laplogger_control" {
		t.Errorf("PostgresDB = %q; want %q", cfg.PostgresDB, "laplogger_control")
	}
	if cfg.BackendPort != "8080" {
		t.Errorf("BackendPort = %q; want %q", cfg.BackendPort, "8080")
	}
}

func TestLoad_OverridesFromEnv(t *testing.T) {
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpass")
	t.Setenv("POSTGRES_HOST", "db.example.com")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_DB", "test_db")
	t.Setenv("JWT_SECRET", "supersecret")
	t.Setenv("BACKEND_PORT", "9090")
	t.Setenv("GOOGLE_CLIENT_ID", "client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "client-secret")
	t.Setenv("GOOGLE_REDIRECT_URI", "http://localhost/callback")

	cfg := Load()

	if cfg.PostgresUser != "testuser" {
		t.Errorf("PostgresUser = %q; want %q", cfg.PostgresUser, "testuser")
	}
	if cfg.PostgresHost != "db.example.com" {
		t.Errorf("PostgresHost = %q; want %q", cfg.PostgresHost, "db.example.com")
	}
	if cfg.JWTSecret != "supersecret" {
		t.Errorf("JWTSecret = %q; want %q", cfg.JWTSecret, "supersecret")
	}
	if cfg.GoogleClientID != "client-id" {
		t.Errorf("GoogleClientID = %q; want %q", cfg.GoogleClientID, "client-id")
	}
	if cfg.GoogleRedirectURI != "http://localhost/callback" {
		t.Errorf("GoogleRedirectURI = %q; want %q", cfg.GoogleRedirectURI, "http://localhost/callback")
	}
}
