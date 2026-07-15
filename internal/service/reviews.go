package service

import (
	"barterswap/internal/apperr"
	"barterswap/internal/domain"
	"context"
	"fmt"
)

// ReviewRepository définit le contrat de persistance dont ReviewService a besoin.
type ReviewRepository interface {
	GetExchangeByID(ctx context.Context, id int) (domain.Exchange, error)
	CreateReview(ctx context.Context, r domain.Review) (domain.Review, error)
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	GetServiceByID(ctx context.Context, id int) (domain.Service, error)
	ListReviewsByTarget(ctx context.Context, targetID int) ([]domain.Review, error)
	ListReviewsByService(ctx context.Context, serviceID int) ([]domain.Review, error)
}

// ReviewService contient la logique métier des évaluations post-échange.
type ReviewService struct {
	repo ReviewRepository
}

// NewReviewService instancie le service métier des avis.
func NewReviewService(repo ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

// CreateReviewInput contient la note et le commentaire d'un avis.
type CreateReviewInput struct {
	Note        int
	Commentaire string
}

// Create enregistre un avis sur un échange terminé (un seul avis par auteur).
func (s *ReviewService) Create(ctx context.Context, authorID, exchangeID int, in CreateReviewInput) (domain.Review, error) {
	if in.Note < 1 || in.Note > 5 {
		return domain.Review{}, fmt.Errorf("%w: la note doit être comprise entre 1 et 5", apperr.ErrValidation)
	}
	ex, err := s.repo.GetExchangeByID(ctx, exchangeID)
	if err != nil {
		return domain.Review{}, err
	}
	if ex.RequesterID != authorID && ex.OwnerID != authorID {
		return domain.Review{}, apperr.ErrForbidden
	}
	if ex.Status != domain.StatusCompleted {
		return domain.Review{}, fmt.Errorf("%w: seul un échange terminé peut être évalué", apperr.ErrValidation)
	}
	targetID := ex.RequesterID
	if authorID == ex.RequesterID {
		targetID = ex.OwnerID
	}
	return s.repo.CreateReview(ctx, domain.Review{
		ExchangeID:  exchangeID,
		AuthorID:    authorID,
		TargetID:    targetID,
		Note:        in.Note,
		Commentaire: in.Commentaire,
	})
}

// ListByUser retourne les avis reçus par un utilisateur.
func (s *ReviewService) ListByUser(ctx context.Context, userID int) ([]domain.Review, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.repo.ListReviewsByTarget(ctx, userID)
}

// ListByService retourne les avis liés au prestataire d'un service.
func (s *ReviewService) ListByService(ctx context.Context, serviceID int) ([]domain.Review, error) {
	if _, err := s.repo.GetServiceByID(ctx, serviceID); err != nil {
		return nil, err
	}
	return s.repo.ListReviewsByService(ctx, serviceID)
}
