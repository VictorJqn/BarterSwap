package repository

import (
	"barterswap/internal/apperr"
	"barterswap/internal/domain"
	"context"
	"fmt"
	"time"
)

func (s *Store) CreateReview(ctx context.Context, r domain.Review) (domain.Review, error) {
	var createdAt time.Time
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO reviews (exchange_id, author_id, target_id, note, commentaire)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		r.ExchangeID, r.AuthorID, r.TargetID, r.Note, r.Commentaire,
	).Scan(&r.ID, &createdAt)
	if err != nil {

		if isUniqueViolation(err) {
			return domain.Review{}, fmt.Errorf("%w: un avis a déjà été déposé pour cet échange", apperr.ErrValidation)
		}
		return domain.Review{}, err
	}
	r.CreatedAt = createdAt.Format(time.RFC3339)
	return r, nil
}

func (s *Store) ListReviewsByTarget(ctx context.Context, targetID int) ([]domain.Review, error) {
	return s.listReviews(ctx,
		`SELECT id, exchange_id, author_id, target_id, note, commentaire, created_at
		 FROM reviews
		 WHERE target_id = $1
		 ORDER BY created_at DESC`, targetID)
}

// Avis portant sur le prestataire du service (cible = offreur), tous échanges confondus.
func (s *Store) ListReviewsByService(ctx context.Context, serviceID int) ([]domain.Review, error) {
	return s.listReviews(ctx,
		`SELECT r.id, r.exchange_id, r.author_id, r.target_id, r.note, r.commentaire, r.created_at
		 FROM reviews r
		 JOIN exchanges e ON e.id = r.exchange_id
		 WHERE e.service_id = $1 AND r.target_id = e.owner_id
		 ORDER BY r.created_at DESC`, serviceID)
}

func (s *Store) listReviews(ctx context.Context, query string, arg int) ([]domain.Review, error) {
	rows, err := s.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := []domain.Review{}
	for rows.Next() {
		var (
			r         domain.Review
			createdAt time.Time
		)
		if err := rows.Scan(&r.ID, &r.ExchangeID, &r.AuthorID, &r.TargetID, &r.Note, &r.Commentaire, &createdAt); err != nil {
			return nil, err
		}
		r.CreatedAt = createdAt.Format(time.RFC3339)
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}
