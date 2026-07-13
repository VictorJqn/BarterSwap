package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *sqlStore) CreateService(ctx context.Context, svc Service) (Service, error) {
	var createdAt time.Time
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO services (provider_id, titre, description, categorie, duree_minutes, credits, ville, actif)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at`,
		svc.ProviderID, svc.Titre, svc.Description, svc.Categorie,
		svc.DureeMinutes, svc.Credits, svc.Ville, svc.Actif,
	).Scan(&svc.ID, &createdAt)
	if err != nil {
		return Service{}, err
	}
	svc.CreatedAt = createdAt.Format(time.RFC3339)
	return svc, nil
}

func (s *sqlStore) GetServiceByID(ctx context.Context, id int) (Service, error) {
	var (
		svc       Service
		createdAt time.Time
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT id, provider_id, titre, description, categorie, duree_minutes, credits, ville, actif, created_at
		 FROM services WHERE id = $1`, id,
	).Scan(&svc.ID, &svc.ProviderID, &svc.Titre, &svc.Description, &svc.Categorie,
		&svc.DureeMinutes, &svc.Credits, &svc.Ville, &svc.Actif, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Service{}, fmt.Errorf("%w: service %d", ErrNotFound, id)
	}
	if err != nil {
		return Service{}, err
	}
	svc.CreatedAt = createdAt.Format(time.RFC3339)
	return svc, nil
}

func (s *sqlStore) ListServices(ctx context.Context, filter ServiceFilter) ([]Service, error) {
	var (
		conds  []string
		args   []any
		argIdx = 1
	)

	if filter.Categorie != "" {
		conds = append(conds, fmt.Sprintf("categorie = $%d", argIdx))
		args = append(args, filter.Categorie)
		argIdx++
	}
	if filter.Ville != "" {
		conds = append(conds, fmt.Sprintf("ville = $%d", argIdx))
		args = append(args, filter.Ville)
		argIdx++
	}
	if filter.Search != "" {
		conds = append(conds, fmt.Sprintf("(titre ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	query := `SELECT id, provider_id, titre, description, categorie, duree_minutes, credits, ville, actif, created_at
	          FROM services`
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []Service{}
	for rows.Next() {
		var (
			svc       Service
			createdAt time.Time
		)
		if err := rows.Scan(&svc.ID, &svc.ProviderID, &svc.Titre, &svc.Description, &svc.Categorie,
			&svc.DureeMinutes, &svc.Credits, &svc.Ville, &svc.Actif, &createdAt); err != nil {
			return nil, err
		}
		svc.CreatedAt = createdAt.Format(time.RFC3339)
		services = append(services, svc)
	}
	return services, rows.Err()
}

func (s *sqlStore) UpdateService(ctx context.Context, svc Service) (Service, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE services
		 SET titre = $1, description = $2, categorie = $3, duree_minutes = $4,
		     credits = $5, ville = $6, actif = $7
		 WHERE id = $8`,
		svc.Titre, svc.Description, svc.Categorie, svc.DureeMinutes,
		svc.Credits, svc.Ville, svc.Actif, svc.ID,
	)
	if err != nil {
		return Service{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return Service{}, fmt.Errorf("%w: service %d", ErrNotFound, svc.ID)
	}
	return s.GetServiceByID(ctx, svc.ID)
}

func (s *sqlStore) DeleteService(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM services WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%w: service %d", ErrNotFound, id)
	}
	return nil
}
