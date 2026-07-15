package config

import "os"

// Config regroupe la configuration de l'application.
type Config struct {
	DatabaseURL string
	Port        string
}

// DefaultDatabaseURL est la DSN PostgreSQL par défaut.
const DefaultDatabaseURL = "postgres://barterswap:barterswap@localhost:5434/barterswap?sslmode=disable"

// Load lit la configuration depuis les variables d'environnement.
func Load() Config {
	return Config{
		DatabaseURL: EnvOr("DATABASE_URL", DefaultDatabaseURL),
		Port:        EnvOr("PORT", "8080"),
	}
}

// EnvOr retourne la variable d'environnement ou la valeur par défaut.
func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
