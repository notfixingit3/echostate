package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	// Unset all relevant env vars to test defaults.
	for _, key := range []string{
		"ECHOSTATE_ENV", "ECHOSTATE_PORT", "DATABASE_URL",
		"BROWSER_WS_URL",
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB",
	} {
		t.Setenv(key, "")
	}

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, "development", cfg.Env)
	require.Equal(t, "8080", cfg.Port)
	require.Equal(t, "ws://localhost:3000/", cfg.BrowserWSURL)
	require.Equal(t, "postgres://echostate:echostate@localhost:5432/echostate?sslmode=disable", cfg.DatabaseURL)
}

func TestLoad_Overrides(t *testing.T) {
	// Clear DATABASE_URL so POSTGRES_* vars are used to build the URL.
	t.Setenv("DATABASE_URL", "")
	t.Setenv("ECHOSTATE_ENV", "staging")
	t.Setenv("ECHOSTATE_PORT", "9090")
	t.Setenv("BROWSER_WS_URL", "ws://browser:3000/")
	t.Setenv("POSTGRES_HOST", "pg.example.com")
	t.Setenv("POSTGRES_PORT", "15432")
	t.Setenv("POSTGRES_USER", "admin")
	t.Setenv("POSTGRES_PASSWORD", "secret")
	t.Setenv("POSTGRES_DB", "echostate_prod")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, "staging", cfg.Env)
	require.Equal(t, "9090", cfg.Port)
	require.Equal(t, "ws://browser:3000/", cfg.BrowserWSURL)
	require.Equal(t, "postgres://admin:secret@pg.example.com:15432/echostate_prod?sslmode=disable", cfg.DatabaseURL)
}

func TestLoad_DatabaseURLPrecedence(t *testing.T) {
	// When DATABASE_URL is set, POSTGRES_* vars must be ignored.
	t.Setenv("DATABASE_URL", "postgres://custom:custom@customhost:5555/customdb?sslmode=disable")
	t.Setenv("POSTGRES_HOST", "ignored-host")
	t.Setenv("POSTGRES_PORT", "ignored-port")
	t.Setenv("POSTGRES_USER", "ignored-user")
	t.Setenv("POSTGRES_PASSWORD", "ignored-pass")
	t.Setenv("POSTGRES_DB", "ignored-db")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, "postgres://custom:custom@customhost:5555/customdb?sslmode=disable", cfg.DatabaseURL)
}

func TestGetEnv_ReturnsValueWhenSet(t *testing.T) {
	t.Setenv("TEST_GETENV_SET", "hello")
	got := getEnv("TEST_GETENV_SET", "fallback")
	require.Equal(t, "hello", got)
}

func TestGetEnv_ReturnsFallbackWhenUnset(t *testing.T) {
	// Ensure the env var is not set.
	t.Setenv("TEST_GETENV_UNSET", "")
	got := getEnv("TEST_GETENV_UNSET", "fallback")
	require.Equal(t, "fallback", got)
}

func TestGetEnv_ReturnsFallbackWhenEmpty(t *testing.T) {
	t.Setenv("TEST_GETENV_EMPTY", "")
	got := getEnv("TEST_GETENV_EMPTY", "fallback")
	require.Equal(t, "fallback", got)
}

func TestGetEnv_EmptyStringIsNotAValue(t *testing.T) {
	// Verify that an explicitly empty string is treated as unset.
	t.Setenv("TEST_GETENV_EMPTY2", "")
	got := getEnv("TEST_GETENV_EMPTY2", "default")
	require.Equal(t, "default", got)
}

func TestLoad_DefaultsUnchangedByOtherEnvVars(t *testing.T) {
	// Set unrelated env vars to ensure they don't leak into config.
	t.Setenv("PATH", "/custom/path")
	t.Setenv("HOME", "/custom/home")

	for _, key := range []string{
		"ECHOSTATE_ENV", "ECHOSTATE_PORT", "DATABASE_URL",
		"BROWSER_WS_URL",
		"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB",
	} {
		t.Setenv(key, "")
	}

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "development", cfg.Env)
	require.Equal(t, "8080", cfg.Port)
}
