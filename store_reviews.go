package main

import (
	"context"
	"fmt"
	"time"
)

func (s *sqlStore) CreateReview(ctx context.Context, r Review) (Review, error) {
	var createdAt time.Time
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO reviews (exchange_id, author_id, target_id, note, commentaire)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		r.ExchangeID, r.AuthorID, r.TargetID, r.Note, r.Commentaire,
	).Scan(&r.ID, &createdAt)
	if err != nil {
		
		if isUniqueViolation(err) {
			return Review{}, fmt.Errorf("%w: un avis a déjà été déposé pour cet échange", ErrValidation)
		}
		return Review{}, err
	}
	r.CreatedAt = createdAt.Format(time.RFC3339)
	return r, nil
}
