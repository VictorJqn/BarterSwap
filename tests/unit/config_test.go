package unit

import (
	"testing"

	"barterswap/internal/config"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PORT", "")

	cfg := config.Load()
	if cfg.DatabaseURL != config.DefaultDatabaseURL {
		t.Fatalf("databaseURL = %q, want default", cfg.DatabaseURL)
	}
	if cfg.Port != "8080" {
		t.Fatalf("port = %q, want 8080", cfg.Port)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://custom/db")
	t.Setenv("PORT", "9090")

	cfg := config.Load()
	if cfg.DatabaseURL != "postgres://custom/db" {
		t.Fatalf("databaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.Port != "9090" {
		t.Fatalf("port = %q", cfg.Port)
	}
}

func TestEnvOr(t *testing.T) {
	t.Setenv("TEST_KEY", "value")
	if got := config.EnvOr("TEST_KEY", "fallback"); got != "value" {
		t.Fatalf("config.EnvOr with env = %q, want value", got)
	}
	if got := config.EnvOr("MISSING_KEY", "fallback"); got != "fallback" {
		t.Fatalf("config.EnvOr without env = %q, want fallback", got)
	}
}
