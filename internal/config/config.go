package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration loaded from the environment.
type Config struct {
	Env                string
	Port               string
	DatabaseURL        string
	BrowserWSURL       string
	FrontendURL        string
	PwhoisEnabled      bool
	PwhoisCacheTTLHours int
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	pwhoisEnabled, _ := strconv.ParseBool(getEnv("ECHOSTATE_PWHOIS_ENABLED", "true"))
	pwhoisCacheTTL, _ := strconv.Atoi(getEnv("PWHOIS_CACHE_TTL_HOURS", "24"))

	cfg := &Config{
		Env:                getEnv("ECHOSTATE_ENV", "development"),
		Port:               getEnv("ECHOSTATE_PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		BrowserWSURL:       getEnv("BROWSER_WS_URL", "ws://localhost:3000/"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		PwhoisEnabled:      pwhoisEnabled,
		PwhoisCacheTTLHours: pwhoisCacheTTL,
	}

	if cfg.DatabaseURL == "" {
		host := getEnv("POSTGRES_HOST", "localhost")
		port := getEnv("POSTGRES_PORT", "5432")
		user := getEnv("POSTGRES_USER", "echostate")
		pass := getEnv("POSTGRES_PASSWORD", "echostate")
		dbname := getEnv("POSTGRES_DB", "echostate")

		cfg.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname)
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL or PostgreSQL env vars are required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
