package main

import (
	"context"
	"fmt"
)

type reviewRepository interface {
	GetExchangeByID(ctx context.Context, id int) (Exchange, error)
	CreateReview(ctx context.Context, r Review) (Review, error)
	GetUserByID(ctx context.Context, id int) (User, error)
	GetServiceByID(ctx context.Context, id int) (Service, error)
	ListReviewsByTarget(ctx context.Context, targetID int) ([]Review, error)
	ListReviewsByService(ctx context.Context, serviceID int) ([]Review, error)
}

type ReviewService struct {
	repo reviewRepository
}

func NewReviewService(repo reviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

type CreateReviewInput struct {
	Note        int
	Commentaire string
}

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

func (s *ReviewService) ListByUser(ctx context.Context, userID int) ([]Review, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.repo.ListReviewsByTarget(ctx, userID)
}

func (s *ReviewService) ListByService(ctx context.Context, serviceID int) ([]Review, error) {
	if _, err := s.repo.GetServiceByID(ctx, serviceID); err != nil {
		return nil, err
	}
	return s.repo.ListReviewsByService(ctx, serviceID)
}
