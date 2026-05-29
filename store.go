package main

import (
	"context"
	"database/sql"
)

type Store interface {
	CreateUser(ctx context.Context, pseudo, bio, ville string) (User, error)
	GetUserByID(ctx context.Context, id int) (User, error)
	UpdateUser(ctx context.Context, id int, pseudo, bio, ville string) (User, error)
	GetSkills(ctx context.Context, userID int) ([]Skill, error)
	ReplaceSkills(ctx context.Context, userID int, skills []Skill) error
}

type sqlStore struct {
	db *sql.DB
}

func newSQLStore(db *sql.DB) *sqlStore {
	return &sqlStore{db: db}
}

var _ Store = (*sqlStore)(nil)
