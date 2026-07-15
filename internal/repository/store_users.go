package repository

import (
	"barterswap/internal/apperr"
	"barterswap/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *Store) CreateUser(ctx context.Context, pseudo, bio, ville string) (domain.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.User{}, err
	}
	defer tx.Rollback()

	var (
		u         domain.User
		createdAt time.Time
	)
	err = tx.QueryRowContext(ctx,
		`INSERT INTO users (pseudo, bio, ville)
		 VALUES ($1, $2, $3)
		 RETURNING id, pseudo, bio, ville, created_at`,
		pseudo, bio, ville,
	).Scan(&u.ID, &u.Pseudo, &u.Bio, &u.Ville, &createdAt)
	if err != nil {
		return domain.User{}, err
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO credit_transactions (user_id, montant, type)
		 VALUES ($1, $2, $3)`,
		u.ID, domain.WelcomeCredits, domain.TxEarn,
	); err != nil {
		return domain.User{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.User{}, err
	}

	u.CreatedAt = createdAt.Format(time.RFC3339)
	u.CreditBalance = domain.WelcomeCredits
	u.Skills = []domain.Skill{}
	return u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	var (
		u         domain.User
		createdAt time.Time
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT id, pseudo, bio, ville, created_at,
		        COALESCE((SELECT SUM(montant) FROM credit_transactions WHERE user_id = users.id), 0)
		 FROM users
		 WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Pseudo, &u.Bio, &u.Ville, &createdAt, &u.CreditBalance)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, fmt.Errorf("%w: utilisateur %d", apperr.ErrNotFound, id)
	}
	if err != nil {
		return domain.User{}, err
	}

	skills, err := s.GetSkills(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u.CreatedAt = createdAt.Format(time.RFC3339)
	u.Skills = skills
	return u, nil
}

func (s *Store) UpdateUser(ctx context.Context, id int, pseudo, bio, ville string) (domain.User, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE users SET pseudo = $1, bio = $2, ville = $3 WHERE id = $4`,
		pseudo, bio, ville, id,
	)
	if err != nil {
		return domain.User{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.User{}, fmt.Errorf("%w: utilisateur %d", apperr.ErrNotFound, id)
	}
	return s.GetUserByID(ctx, id)
}

func (s *Store) GetSkills(ctx context.Context, userID int) ([]domain.Skill, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT nom, niveau FROM skills WHERE user_id = $1 ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := []domain.Skill{}
	for rows.Next() {
		var sk domain.Skill
		if err := rows.Scan(&sk.Nom, &sk.Niveau); err != nil {
			return nil, err
		}
		skills = append(skills, sk)
	}
	return skills, rows.Err()
}

func (s *Store) ReplaceSkills(ctx context.Context, userID int, skills []domain.Skill) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID,
	).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: utilisateur %d", apperr.ErrNotFound, userID)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM skills WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, sk := range skills {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO skills (user_id, nom, niveau) VALUES ($1, $2, $3)`,
			userID, sk.Nom, sk.Niveau,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}
