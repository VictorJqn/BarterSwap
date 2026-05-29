package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func openDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("ouverture de la base : %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("connexion à la base : %w", err)
	}
	return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("création du schéma : %w", err)
	}
	return nil
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    pseudo     TEXT        NOT NULL,
    bio        TEXT        NOT NULL DEFAULT '',
    ville      TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS skills (
    id      SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nom     TEXT    NOT NULL,
    niveau  TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_skills_user ON skills(user_id);

CREATE TABLE IF NOT EXISTS services (
    id            SERIAL PRIMARY KEY,
    provider_id   INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    titre         TEXT        NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    categorie     TEXT        NOT NULL,
    duree_minutes INTEGER     NOT NULL DEFAULT 0,
    credits       INTEGER     NOT NULL DEFAULT 0,
    ville         TEXT        NOT NULL DEFAULT '',
    actif         BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_services_categorie ON services(categorie);
CREATE INDEX IF NOT EXISTS idx_services_ville ON services(ville);

CREATE TABLE IF NOT EXISTS exchanges (
    id           SERIAL PRIMARY KEY,
    service_id   INTEGER     NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    requester_id INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    owner_id     INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       TEXT        NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Garantit qu'un service n'a qu'un seul échange actif (pending/accepted) à la
-- fois : toute violation remonte naturellement en conflit (409).
CREATE UNIQUE INDEX IF NOT EXISTS uq_exchanges_active_service
    ON exchanges(service_id) WHERE status IN ('pending', 'accepted');

CREATE TABLE IF NOT EXISTS credit_transactions (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exchange_id INTEGER     REFERENCES exchanges(id) ON DELETE SET NULL,
    montant     INTEGER     NOT NULL,
    type        TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_credit_tx_user ON credit_transactions(user_id);

CREATE TABLE IF NOT EXISTS reviews (
    id          SERIAL PRIMARY KEY,
    exchange_id INTEGER     NOT NULL REFERENCES exchanges(id) ON DELETE CASCADE,
    author_id   INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id   INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    note        INTEGER     NOT NULL CHECK (note BETWEEN 1 AND 5),
    commentaire TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (exchange_id, author_id)
);
`
