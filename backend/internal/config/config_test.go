package config

import (
	"testing"
	"time"
)

func TestLoadWithValidEnvironment(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("JWT_SECRET", "this-is-a-secret-with-at-least-32-chars")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != "8080" {
		t.Fatalf("port = %q, want default 8080", cfg.Port)
	}
	if cfg.JWTTTL != 24*time.Hour {
		t.Fatalf("ttl = %v, want 24h", cfg.JWTTTL)
	}
	if cfg.Environment != "development" {
		t.Fatalf("environment = %q, want development", cfg.Environment)
	}
}

func TestLoadRejectsMissingDatabaseURL(t *testing.T) {
	t.Setenv("JWT_SECRET", "this-is-a-secret-with-at-least-32-chars")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoadRejectsShortJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("JWT_SECRET", "short")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when JWT_SECRET is too short")
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("PORT", "not-a-port")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("JWT_SECRET", "this-is-a-secret-with-at-least-32-chars")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when PORT is invalid")
	}
}

func TestLoadRejectsInvalidTTL(t *testing.T) {
	t.Setenv("JWT_TTL_HOURS", "-5")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("JWT_SECRET", "this-is-a-secret-with-at-least-32-chars")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when JWT_TTL_HOURS is invalid")
	}
}
