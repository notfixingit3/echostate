package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	Neo4jURI           string
	Neo4jUser          string
	Neo4jPassword      string
	Branch             string
	GitHubRepo         string
	UpdateCheckEnabled bool
	LogLevel           string
	LogFormat          string
	MetricsPort        int
	OTelEndpoint       string
	OTelSampler        string
	OTelSamplerArg     string
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	pwhoisEnabled, _ := strconv.ParseBool(getEnv("ECHOSTATE_PWHOIS_ENABLED", "true"))
	pwhoisCacheTTL, _ := strconv.Atoi(getEnv("PWHOIS_CACHE_TTL_HOURS", "24"))
	updateCheckEnabled, _ := strconv.ParseBool(getEnv("ECHOSTATE_UPDATE_CHECK", "true"))
	metricsPort, _ := strconv.Atoi(getEnv("ECHOSTATE_METRICS_PORT", "9090"))

	env := getEnv("ECHOSTATE_ENV", "development")
	defaultSamplerArg := "0.1"
	if env != "production" {
		defaultSamplerArg = "1.0"
	}

	cfg := &Config{
		Env:                env,
		Port:               getEnv("ECHOSTATE_PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		BrowserWSURL:       getEnv("BROWSER_WS_URL", "ws://localhost:3000/"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3001"),
		PwhoisEnabled:      pwhoisEnabled,
		PwhoisCacheTTLHours: pwhoisCacheTTL,
		Neo4jURI:           getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser:          getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword:      getEnv("NEO4J_PASSWORD", "echostate123"),
		Branch:             getEnv("ECHOSTATE_BRANCH", "dev"),
		GitHubRepo:         getEnv("ECHOSTATE_GITHUB_REPO", "notfixingit3/echostate"),
		UpdateCheckEnabled: updateCheckEnabled,
		LogLevel:           getEnv("ECHOSTATE_LOG_LEVEL", "info"),
		LogFormat:          getEnv("ECHOSTATE_LOG_FORMAT", "json"),
		MetricsPort:        metricsPort,
		OTelEndpoint:       os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		OTelSampler:        getEnv("OTEL_TRACES_SAMPLER", "parentbased_traceidratio"),
		OTelSamplerArg:     getEnv("OTEL_TRACES_SAMPLER_ARG", defaultSamplerArg),
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

	if cfg.Env == "production" {
		pepper := strings.TrimSpace(os.Getenv("ECHOSTATE_AUTH_PEPPER"))
		if pepper == "" || pepper == "echostate-dev-pepper" {
			return nil, fmt.Errorf("ECHOSTATE_AUTH_PEPPER must be set to a strong random value in production")
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
