package main

import (
	"context"
	"fmt"
)

// exchangeRepository définit le contrat de persistance dont ExchangeService a besoin.
type exchangeRepository interface {
	GetServiceByID(ctx context.Context, id int) (Service, error)
	CreateExchange(ctx context.Context, e Exchange) (Exchange, error)
	GetExchangeByID(ctx context.Context, id int) (Exchange, error)
	ListExchanges(ctx context.Context, userID int, status string) ([]Exchange, error)
	AcceptExchange(ctx context.Context, exchangeID int, credits int) error
	RejectExchange(ctx context.Context, exchangeID int) error
	CompleteExchange(ctx context.Context, exchangeID int, credits int) error
	CancelExchange(ctx context.Context, exchangeID int, credits int, wasAccepted bool) error
	GetCreditBalance(ctx context.Context, userID int) (int, error)
}

// ExchangeService contient la logique métier du système d'échange.
type ExchangeService struct {
	repo exchangeRepository
}

func NewExchangeService(repo exchangeRepository) *ExchangeService {
	return &ExchangeService{repo: repo}
}

type CreateExchangeInput struct {
	ServiceID int
}

func (s *ExchangeService) Create(ctx context.Context, requesterID int, in CreateExchangeInput) (Exchange, error) {
	if in.ServiceID <= 0 {
		return Exchange{}, fmt.Errorf("%w: identifiant de service invalide", ErrValidation)
	}
	svc, err := s.repo.GetServiceByID(ctx, in.ServiceID)
	if err != nil {
		return Exchange{}, err
	}
	if !svc.Actif {
		return Exchange{}, fmt.Errorf("%w: service inactif", ErrValidation)
	}
	if svc.ProviderID == requesterID {
		return Exchange{}, fmt.Errorf("%w: impossible de demander son propre service", ErrValidation)
	}
	balance, err := s.repo.GetCreditBalance(ctx, requesterID)
	if err != nil {
		return Exchange{}, err
	}
	if balance < svc.Credits {
		return Exchange{}, ErrInsufficientCredits
	}
	return s.repo.CreateExchange(ctx, Exchange{
		ServiceID:   svc.ID,
		RequesterID: requesterID,
		OwnerID:     svc.ProviderID,
	})
}

func (s *ExchangeService) GetByID(ctx context.Context, actorID, exchangeID int) (Exchange, error) {
	ex, err := s.repo.GetExchangeByID(ctx, exchangeID)
	if err != nil {
		return Exchange{}, err
	}
	if ex.RequesterID != actorID && ex.OwnerID != actorID {
		return Exchange{}, ErrForbidden
	}
	return ex, nil
}

func (s *ExchangeService) List(ctx context.Context, actorID int, status string) ([]Exchange, error) {
	if status != "" && !validStatus(status) {
		return nil, fmt.Errorf("%w: statut %q invalide", ErrValidation, status)
	}
	return s.repo.ListExchanges(ctx, actorID, status)
}

func (s *ExchangeService) Accept(ctx context.Context, actorID, exchangeID int) (Exchange, error) {
	ex, svc, err := s.loadExchangeWithService(ctx, exchangeID)
	if err != nil {
		return Exchange{}, err
	}
	if ex.OwnerID != actorID {
		return Exchange{}, ErrForbidden
	}
	if err := s.repo.AcceptExchange(ctx, exchangeID, svc.Credits); err != nil {
		return Exchange{}, err
	}
	return s.repo.GetExchangeByID(ctx, exchangeID)
}

func (s *ExchangeService) Reject(ctx context.Context, actorID, exchangeID int) (Exchange, error) {
	ex, err := s.repo.GetExchangeByID(ctx, exchangeID)
	if err != nil {
		return Exchange{}, err
	}
	if ex.OwnerID != actorID {
		return Exchange{}, ErrForbidden
	}
	if err := s.repo.RejectExchange(ctx, exchangeID); err != nil {
		return Exchange{}, err
	}
	return s.repo.GetExchangeByID(ctx, exchangeID)
}

func (s *ExchangeService) Complete(ctx context.Context, actorID, exchangeID int) (Exchange, error) {
	ex, svc, err := s.loadExchangeWithService(ctx, exchangeID)
	if err != nil {
		return Exchange{}, err
	}
	if ex.OwnerID != actorID {
		return Exchange{}, ErrForbidden
	}
	if err := s.repo.CompleteExchange(ctx, exchangeID, svc.Credits); err != nil {
		return Exchange{}, err
	}
	return s.repo.GetExchangeByID(ctx, exchangeID)
}

func (s *ExchangeService) Cancel(ctx context.Context, actorID, exchangeID int) (Exchange, error) {
	ex, svc, err := s.loadExchangeWithService(ctx, exchangeID)
	if err != nil {
		return Exchange{}, err
	}
	if ex.RequesterID != actorID && ex.OwnerID != actorID {
		return Exchange{}, ErrForbidden
	}
	wasAccepted := ex.Status == StatusAccepted
	if err := s.repo.CancelExchange(ctx, exchangeID, svc.Credits, wasAccepted); err != nil {
		return Exchange{}, err
	}
	return s.repo.GetExchangeByID(ctx, exchangeID)
}

func (s *ExchangeService) loadExchangeWithService(ctx context.Context, exchangeID int) (Exchange, Service, error) {
	ex, err := s.repo.GetExchangeByID(ctx, exchangeID)
	if err != nil {
		return Exchange{}, Service{}, err
	}
	svc, err := s.repo.GetServiceByID(ctx, ex.ServiceID)
	if err != nil {
		return Exchange{}, Service{}, err
	}
	return ex, svc, nil
}
