package main

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PORT", "")

	cfg := loadConfig()
	if cfg.databaseURL != defaultDatabaseURL {
		t.Fatalf("databaseURL = %q, want default", cfg.databaseURL)
	}
	if cfg.port != "8080" {
		t.Fatalf("port = %q, want 8080", cfg.port)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://custom/db")
	t.Setenv("PORT", "9090")

	cfg := loadConfig()
	if cfg.databaseURL != "postgres://custom/db" {
		t.Fatalf("databaseURL = %q", cfg.databaseURL)
	}
	if cfg.port != "9090" {
		t.Fatalf("port = %q", cfg.port)
	}
}

func TestEnvOr(t *testing.T) {
	t.Setenv("TEST_KEY", "value")
	if got := envOr("TEST_KEY", "fallback"); got != "value" {
		t.Fatalf("envOr with env = %q, want value", got)
	}
	if got := envOr("MISSING_KEY", "fallback"); got != "fallback" {
		t.Fatalf("envOr without env = %q, want fallback", got)
	}
}
