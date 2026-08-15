package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the runtime settings for the API service. It is loaded
// once at startup from the environment and validated eagerly so that a
// misconfigured deployment fails fast instead of at request time.
type Config struct {
	Environment string
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
}

const (
	defaultPort     = "8080"
	defaultTTL      = 24 * time.Hour
	environmentDev  = "development"
	environmentProd = "production"
	environmentTest = "test"
)

// Load reads the configuration from the process environment. If a .env
// file exists it is loaded first; explicit environment variables always
// take precedence.
func Load() (Config, error) {
	_ = godotenv.Load()

	port := getenv("PORT", defaultPort)
	if !validPort(port) {
		return Config{}, fmt.Errorf("invalid PORT: %q", port)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	ttlHours := getenv("JWT_TTL_HOURS", "24")
	ttl, err := parseTTL(ttlHours)
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_TTL_HOURS: %w", err)
	}

	environment := getenv("ENVIRONMENT", environmentDev)
	switch environment {
	case environmentDev, environmentProd, environmentTest:
	default:
		return Config{}, fmt.Errorf("invalid ENVIRONMENT: %q (allowed: development, production, test)", environment)
	}

	return Config{
		Environment: environment,
		Port:        port,
		DatabaseURL: databaseURL,
		JWTSecret:   secret,
		JWTTTL:      ttl,
	}, nil
}

func validPort(port string) bool {
	n, err := strconv.Atoi(port)
	return err == nil && n >= 1 && n <= 65535
}

func parseTTL(value string) (time.Duration, error) {
	hours, err := strconv.Atoi(value)
	if err != nil || hours <= 0 {
		return 0, fmt.Errorf("must be a positive number of hours")
	}
	return time.Duration(hours) * time.Hour, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
