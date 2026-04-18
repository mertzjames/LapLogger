package config

import "os"

type Config struct {
	PostgresUser     string
	PostgresPassword string
	PostgresHost     string
	PostgresPort     string
	PostgresDB       string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string
	FrontendURL        string

	JWTSecret string

	BackendPort string
}

func Load() *Config {
	return &Config{
		PostgresUser:       getEnv("POSTGRES_USER", "laplogger"),
		PostgresPassword:   getEnv("POSTGRES_PASSWORD", "laplogger_dev"),
		PostgresHost:       getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:       getEnv("POSTGRES_PORT", "5432"),
		PostgresDB:         getEnv("POSTGRES_DB", "laplogger_control"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURI:  os.Getenv("GOOGLE_REDIRECT_URI"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		BackendPort:        getEnv("BACKEND_PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
