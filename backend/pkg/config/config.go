package config

import (
	"os"
	"strings"
)

type Config struct {
	ServerAddr    string
	DBPath        string
	MigrationsDir string
	SeedDB        bool
}

func Load() Config {
	return Config{
		ServerAddr:    envOrDefault("SERVER_ADDR", ":8080"),
		DBPath:        envOrDefault("DB_PATH", "social-network.db"),
		MigrationsDir: envOrDefault("MIGRATIONS_DIR", "pkg/db/migrations/sqlite"),
		SeedDB:        envBoolOrDefault("SEED_DB", false),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBoolOrDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
