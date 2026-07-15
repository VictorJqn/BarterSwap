package repository

import "database/sql"

// Store est l'implémentation PostgreSQL (couche infrastructure).
type Store struct {
	db *sql.DB
}

// New crée l'accès PostgreSQL utilisé par les services métier.
func New(db *sql.DB) *Store {
	return &Store{db: db}
}
