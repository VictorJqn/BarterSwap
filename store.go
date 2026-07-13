package main

import "database/sql"

// sqlStore est l'implémentation PostgreSQL (couche infrastructure).
// Il satisfait implicitement les interfaces définies par chaque service
// (principe du cours : interfaces petites, côté consommateur).
type sqlStore struct {
	db *sql.DB
}

func newSQLStore(db *sql.DB) *sqlStore {
	return &sqlStore{db: db}
}

// Vérification à la compilation : sqlStore satisfait les contrats des services.
var (
	_ userRepository     = (*sqlStore)(nil)
	_ serviceRepository  = (*sqlStore)(nil)
	_ exchangeRepository = (*sqlStore)(nil)
	_ reviewRepository   = (*sqlStore)(nil)
)
