package main

import (
	"context"
	"testing"
	"time"
)

func TestOpenDBInvalidDSN(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := openDB(ctx, "postgres://invalid:invalid@127.0.0.1:1/nope?sslmode=disable&connect_timeout=1")
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestMigrateCreatesSchema(t *testing.T) {
	store, cleanup := setupIntegration(t)
	defer cleanup()

	ctx := context.Background()
	if _, err := store.GetUserByID(ctx, 999); err == nil {
		t.Fatal("expected not found on empty db")
	}
}
