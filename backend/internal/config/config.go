package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
}

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 { return Config{}, fmt.Errorf("invalid PORT: %q", port) }
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" { return Config{}, fmt.Errorf("DATABASE_URL is required") }
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 { return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters") }
	ttl := 24 * time.Hour
	if value := os.Getenv("JWT_TTL_HOURS"); value != "" {
		hours, err := strconv.Atoi(value)
		if err != nil || hours <= 0 { return Config{}, fmt.Errorf("invalid JWT_TTL_HOURS: %q", value) }
		ttl = time.Duration(hours) * time.Hour
	}
	return Config{Port: port, DatabaseURL: databaseURL, JWTSecret: secret, JWTTTL: ttl}, nil
}
