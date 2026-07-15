package main

import (
	"context"
	"fmt"
)

// reviewRepository définit le contrat de persistance dont ReviewService a besoin.
type reviewRepository interface {
	GetExchangeByID(ctx context.Context, id int) (Exchange, error)
	CreateReview(ctx context.Context, r Review) (Review, error)
	GetUserByID(ctx context.Context, id int) (User, error)
	GetServiceByID(ctx context.Context, id int) (Service, error)
	ListReviewsByTarget(ctx context.Context, targetID int) ([]Review, error)
	ListReviewsByService(ctx context.Context, serviceID int) ([]Review, error)
}

// ReviewService contient la logique métier des évaluations post-échange.
type ReviewService struct {
	repo reviewRepository
}

// NewReviewService instancie le service métier des avis.
func NewReviewService(repo reviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

// CreateReviewInput contient la note et le commentaire d'un avis.
type CreateReviewInput struct {
	Note        int
	Commentaire string
}

// Create enregistre un avis sur un échange terminé (un seul avis par auteur).
func (s *ReviewService) Create(ctx context.Context, authorID, exchangeID int, in CreateReviewInput) (Review, error) {
	if in.Note < 1 || in.Note > 5 {
		return Review{}, fmt.Errorf("%w: la note doit être comprise entre 1 et 5", ErrValidation)
	}
	ex, err := s.repo.GetExchangeByID(ctx, exchangeID)
	if err != nil {
		return Review{}, err
	}
	if ex.RequesterID != authorID && ex.OwnerID != authorID {
		return Review{}, ErrForbidden
	}
	if ex.Status != StatusCompleted {
		return Review{}, fmt.Errorf("%w: seul un échange terminé peut être évalué", ErrValidation)
	}
	targetID := ex.RequesterID
	if authorID == ex.RequesterID {
		targetID = ex.OwnerID
	}
	return s.repo.CreateReview(ctx, Review{
		ExchangeID:  exchangeID,
		AuthorID:    authorID,
		TargetID:    targetID,
		Note:        in.Note,
		Commentaire: in.Commentaire,
	})
}

// ListByUser retourne les avis reçus par un utilisateur.
func (s *ReviewService) ListByUser(ctx context.Context, userID int) ([]Review, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.repo.ListReviewsByTarget(ctx, userID)
}

// ListByService retourne les avis liés au prestataire d'un service.
func (s *ReviewService) ListByService(ctx context.Context, serviceID int) ([]Review, error) {
	if _, err := s.repo.GetServiceByID(ctx, serviceID); err != nil {
		return nil, err
	}
	return s.repo.ListReviewsByService(ctx, serviceID)
}
