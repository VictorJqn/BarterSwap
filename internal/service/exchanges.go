package service

import (
	"barterswap/internal/apperr"
	"barterswap/internal/domain"
	"context"
	"fmt"
)

// ExchangeRepository définit le contrat de persistance dont ExchangeService a besoin.
type ExchangeRepository interface {
	GetServiceByID(ctx context.Context, id int) (domain.Service, error)
	CreateExchange(ctx context.Context, e domain.Exchange) (domain.Exchange, error)
	GetExchangeByID(ctx context.Context, id int) (domain.Exchange, error)
	ListExchanges(ctx context.Context, userID int, status string) ([]domain.Exchange, error)
	AcceptExchange(ctx context.Context, exchangeID int, credits int) error
	RejectExchange(ctx context.Context, exchangeID int) error
	CompleteExchange(ctx context.Context, exchangeID int, credits int) error
	CancelExchange(ctx context.Context, exchangeID int, credits int, wasAccepted bool) error
	GetCreditBalance(ctx context.Context, userID int) (int, error)
}

// ExchangeService contient la logique métier du système d'échange.
type ExchangeService struct {
	repo ExchangeRepository
}

// NewExchangeService instancie le service métier des échanges.
func NewExchangeService(repo ExchangeRepository) *ExchangeService {
	return &ExchangeService{repo: repo}
}

// CreateExchangeInput identifie le service ciblé par une demande d'échange.
type CreateExchangeInput struct {
	ServiceID int
}

// Create enregistre une demande d'échange après validation des règles métier.
func (s *ExchangeService) Create(ctx context.Context, requesterID int, in CreateExchangeInput) (domain.Exchange, error) {
	if in.ServiceID <= 0 {
		return domain.Exchange{}, fmt.Errorf("%w: identifiant de service invalide", apperr.ErrValidation)
	}
	svc, err := s.repo.GetServiceByID(ctx, in.ServiceID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if !svc.Actif {
		return domain.Exchange{}, fmt.Errorf("%w: service inactif", apperr.ErrValidation)
	}
	if svc.ProviderID == requesterID {
		return domain.Exchange{}, fmt.Errorf("%w: impossible de demander son propre service", apperr.ErrValidation)
	}
	balance, err := s.repo.GetCreditBalance(ctx, requesterID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if balance < svc.Credits {
		return domain.Exchange{}, apperr.ErrInsufficientCredits
	}
	return s.repo.CreateExchange(ctx, domain.Exchange{
		ServiceID:   svc.ID,
		RequesterID: requesterID,
		OwnerID:     svc.ProviderID,
	})
}

// GetByID retourne un échange si l'utilisateur connecté en est participant.
func (s *ExchangeService) GetByID(ctx context.Context, actorID, exchangeID int) (domain.Exchange, error) {
	ex, err := s.repo.GetExchangeByID(ctx, exchangeID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if ex.RequesterID != actorID && ex.OwnerID != actorID {
		return domain.Exchange{}, apperr.ErrForbidden
	}
	return ex, nil
}

// List retourne les échanges envoyés et reçus par l'utilisateur, avec filtre optionnel.
func (s *ExchangeService) List(ctx context.Context, actorID int, status string) ([]domain.Exchange, error) {
	if status != "" && !domain.ValidStatus(status) {
		return nil, fmt.Errorf("%w: statut %q invalide", apperr.ErrValidation, status)
	}
	return s.repo.ListExchanges(ctx, actorID, status)
}

// Accept fait passer l'échange en accepted et bloque les crédits du demandeur.
func (s *ExchangeService) Accept(ctx context.Context, actorID, exchangeID int) (domain.Exchange, error) {
	ex, svc, err := s.loadExchangeWithService(ctx, exchangeID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if ex.OwnerID != actorID {
		return domain.Exchange{}, apperr.ErrForbidden
	}
	if err := s.repo.AcceptExchange(ctx, exchangeID, svc.Credits); err != nil {
		return domain.Exchange{}, err
	}
	return s.repo.GetExchangeByID(ctx, exchangeID)
}

// Reject refuse une demande en attente (statut rejected).
func (s *ExchangeService) Reject(ctx context.Context, actorID, exchangeID int) (domain.Exchange, error) {
	ex, err := s.repo.GetExchangeByID(ctx, exchangeID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if ex.OwnerID != actorID {
		return domain.Exchange{}, apperr.ErrForbidden
	}
	if err := s.repo.RejectExchange(ctx, exchangeID); err != nil {
		return domain.Exchange{}, err
	}
	return s.repo.GetExchangeByID(ctx, exchangeID)
}

// Complete termine un échange accepté et crédite l'offreur.
func (s *ExchangeService) Complete(ctx context.Context, actorID, exchangeID int) (domain.Exchange, error) {
	ex, svc, err := s.loadExchangeWithService(ctx, exchangeID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if ex.OwnerID != actorID {
		return domain.Exchange{}, apperr.ErrForbidden
	}
	if err := s.repo.CompleteExchange(ctx, exchangeID, svc.Credits); err != nil {
		return domain.Exchange{}, err
	}
	return s.repo.GetExchangeByID(ctx, exchangeID)
}

// Cancel annule un échange et restitue les crédits si nécessaire.
func (s *ExchangeService) Cancel(ctx context.Context, actorID, exchangeID int) (domain.Exchange, error) {
	ex, svc, err := s.loadExchangeWithService(ctx, exchangeID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if ex.RequesterID != actorID && ex.OwnerID != actorID {
		return domain.Exchange{}, apperr.ErrForbidden
	}
	wasAccepted := ex.Status == domain.StatusAccepted
	if err := s.repo.CancelExchange(ctx, exchangeID, svc.Credits, wasAccepted); err != nil {
		return domain.Exchange{}, err
	}
	return s.repo.GetExchangeByID(ctx, exchangeID)
}

func (s *ExchangeService) loadExchangeWithService(ctx context.Context, exchangeID int) (domain.Exchange, domain.Service, error) {
	ex, err := s.repo.GetExchangeByID(ctx, exchangeID)
	if err != nil {
		return domain.Exchange{}, domain.Service{}, err
	}
	svc, err := s.repo.GetServiceByID(ctx, ex.ServiceID)
	if err != nil {
		return domain.Exchange{}, domain.Service{}, err
	}
	return ex, svc, nil
}
