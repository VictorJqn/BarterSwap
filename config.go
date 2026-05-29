package main

import "os"

type config struct {
	databaseURL string
	port        string
}

const defaultDatabaseURL = "postgres://barterswap:barterswap@localhost:5434/barterswap?sslmode=disable"

func loadConfig() config {
	return config{
		databaseURL: envOr("DATABASE_URL", defaultDatabaseURL),
		port:        envOr("PORT", "8080"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
