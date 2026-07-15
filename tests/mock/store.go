package mock

import (
	"barterswap/internal/apperr"
	"barterswap/internal/domain"
	"context"
	"fmt"
	"strings"
	"time"
)

// Store simule la couche SQL en mémoire pour les tests unitaires et API.
type Store struct {
	users     map[int]domain.User
	skills    map[int][]domain.Skill
	services  map[int]domain.Service
	exchanges map[int]domain.Exchange
	reviews   map[int]domain.Review
	credits   map[int]int
	stats     map[int]domain.UserStats

	nextUserID     int
	nextServiceID  int
	nextExchangeID int
	nextReviewID   int
}

func New() *Store {
	return &Store{
		users:     make(map[int]domain.User),
		skills:    make(map[int][]domain.Skill),
		services:  make(map[int]domain.Service),
		exchanges: make(map[int]domain.Exchange),
		reviews:   make(map[int]domain.Review),
		credits:   make(map[int]int),
		stats:     make(map[int]domain.UserStats),
	}
}

func (m *Store) SeedUser(u domain.User, balance int, skills []domain.Skill) {
	if u.ID == 0 {
		m.nextUserID++
		u.ID = m.nextUserID
	}
	u.CreditBalance = balance
	u.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	m.users[u.ID] = u
	m.credits[u.ID] = balance
	m.skills[u.ID] = skills
}

func (m *Store) SeedService(s domain.Service) {
	if s.ID == 0 {
		m.nextServiceID++
		s.ID = m.nextServiceID
	}
	s.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	m.services[s.ID] = s
}

func (m *Store) SeedExchange(e domain.Exchange) {
	if e.ID == 0 {
		m.nextExchangeID++
		e.ID = m.nextExchangeID
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if e.CreatedAt == "" {
		e.CreatedAt = now
	}
	if e.UpdatedAt == "" {
		e.UpdatedAt = now
	}
	m.exchanges[e.ID] = e
}

func (m *Store) hasActiveExchange(serviceID int) bool {
	for _, ex := range m.exchanges {
		if ex.ServiceID == serviceID && (ex.Status == domain.StatusPending || ex.Status == domain.StatusAccepted) {
			return true
		}
	}
	return false
}

func (m *Store) CreateUser(ctx context.Context, pseudo, bio, ville string) (domain.User, error) {
	m.nextUserID++
	u := domain.User{
		ID:            m.nextUserID,
		Pseudo:        pseudo,
		Bio:           bio,
		Ville:         ville,
		CreditBalance: domain.WelcomeCredits,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		Skills:        []domain.Skill{},
	}
	m.users[u.ID] = u
	m.credits[u.ID] = domain.WelcomeCredits
	m.skills[u.ID] = []domain.Skill{}
	return u, nil
}

func (m *Store) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return domain.User{}, fmt.Errorf("%w: utilisateur %d", apperr.ErrNotFound, id)
	}
	u.CreditBalance = m.credits[id]
	u.Skills = m.skills[id]
	return u, nil
}

func (m *Store) UpdateUser(ctx context.Context, id int, pseudo, bio, ville string) (domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return domain.User{}, fmt.Errorf("%w: utilisateur %d", apperr.ErrNotFound, id)
	}
	u.Pseudo = pseudo
	u.Bio = bio
	u.Ville = ville
	m.users[id] = u
	return m.GetUserByID(ctx, id)
}

func (m *Store) GetSkills(ctx context.Context, userID int) ([]domain.Skill, error) {
	if _, ok := m.users[userID]; !ok {
		return nil, fmt.Errorf("%w: utilisateur %d", apperr.ErrNotFound, userID)
	}
	return m.skills[userID], nil
}

func (m *Store) ReplaceSkills(ctx context.Context, userID int, skills []domain.Skill) error {
	if _, ok := m.users[userID]; !ok {
		return fmt.Errorf("%w: utilisateur %d", apperr.ErrNotFound, userID)
	}
	m.skills[userID] = skills
	return nil
}

func (m *Store) GetUserStats(ctx context.Context, userID int) (domain.UserStats, error) {
	if s, ok := m.stats[userID]; ok {
		s.CreditBalance = m.credits[userID]
		return s, nil
	}
	if _, ok := m.users[userID]; !ok {
		return domain.UserStats{}, fmt.Errorf("%w: utilisateur %d", apperr.ErrNotFound, userID)
	}
	return domain.UserStats{
		UserID:        userID,
		CreditBalance: m.credits[userID],
	}, nil
}

func (m *Store) CreateService(ctx context.Context, s domain.Service) (domain.Service, error) {
	m.nextServiceID++
	s.ID = m.nextServiceID
	s.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	m.services[s.ID] = s
	return s, nil
}

func (m *Store) GetServiceByID(ctx context.Context, id int) (domain.Service, error) {
	s, ok := m.services[id]
	if !ok {
		return domain.Service{}, fmt.Errorf("%w: service %d", apperr.ErrNotFound, id)
	}
	return s, nil
}

func (m *Store) ListServices(ctx context.Context, filter domain.ServiceFilter) ([]domain.Service, error) {
	out := []domain.Service{}
	for _, s := range m.services {
		if filter.Categorie != "" && s.Categorie != filter.Categorie {
			continue
		}
		if filter.Ville != "" && s.Ville != filter.Ville {
			continue
		}
		if filter.Search != "" {
			q := strings.ToLower(filter.Search)
			if !strings.Contains(strings.ToLower(s.Titre), q) &&
				!strings.Contains(strings.ToLower(s.Description), q) {
				continue
			}
		}
		out = append(out, s)
	}
	return out, nil
}

func (m *Store) UpdateService(ctx context.Context, s domain.Service) (domain.Service, error) {
	if _, ok := m.services[s.ID]; !ok {
		return domain.Service{}, fmt.Errorf("%w: service %d", apperr.ErrNotFound, s.ID)
	}
	m.services[s.ID] = s
	return s, nil
}

func (m *Store) DeleteService(ctx context.Context, id int) error {
	if _, ok := m.services[id]; !ok {
		return fmt.Errorf("%w: service %d", apperr.ErrNotFound, id)
	}
	delete(m.services, id)
	return nil
}

func (m *Store) CreateExchange(ctx context.Context, e domain.Exchange) (domain.Exchange, error) {
	if m.hasActiveExchange(e.ServiceID) {
		return domain.Exchange{}, fmt.Errorf("%w: service déjà réservé", apperr.ErrConflict)
	}
	m.nextExchangeID++
	e.ID = m.nextExchangeID
	e.Status = domain.StatusPending
	now := time.Now().UTC().Format(time.RFC3339)
	e.CreatedAt = now
	e.UpdatedAt = now
	m.exchanges[e.ID] = e
	return e, nil
}

func (m *Store) GetExchangeByID(ctx context.Context, id int) (domain.Exchange, error) {
	e, ok := m.exchanges[id]
	if !ok {
		return domain.Exchange{}, fmt.Errorf("%w: échange %d", apperr.ErrNotFound, id)
	}
	return e, nil
}

func (m *Store) ListExchanges(ctx context.Context, userID int, status string) ([]domain.Exchange, error) {
	out := []domain.Exchange{}
	for _, e := range m.exchanges {
		if e.RequesterID != userID && e.OwnerID != userID {
			continue
		}
		if status != "" && e.Status != status {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

func (m *Store) AcceptExchange(ctx context.Context, exchangeID int, credits int) error {
	e, ok := m.exchanges[exchangeID]
	if !ok {
		return fmt.Errorf("%w: échange %d", apperr.ErrNotFound, exchangeID)
	}
	if e.Status != domain.StatusPending {
		return fmt.Errorf("%w: seul un échange en attente peut être accepté", apperr.ErrValidation)
	}
	if m.credits[e.RequesterID] < credits {
		return apperr.ErrInsufficientCredits
	}
	m.credits[e.RequesterID] -= credits
	e.Status = domain.StatusAccepted
	e.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	m.exchanges[exchangeID] = e
	return nil
}

func (m *Store) RejectExchange(ctx context.Context, exchangeID int) error {
	e, ok := m.exchanges[exchangeID]
	if !ok {
		return fmt.Errorf("%w: échange %d", apperr.ErrNotFound, exchangeID)
	}
	if e.Status != domain.StatusPending {
		return fmt.Errorf("%w: seul un échange en attente peut être refusé", apperr.ErrValidation)
	}
	e.Status = domain.StatusRejected
	e.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	m.exchanges[exchangeID] = e
	return nil
}

func (m *Store) CompleteExchange(ctx context.Context, exchangeID int, credits int) error {
	e, ok := m.exchanges[exchangeID]
	if !ok {
		return fmt.Errorf("%w: échange %d", apperr.ErrNotFound, exchangeID)
	}
	if e.Status != domain.StatusAccepted {
		return fmt.Errorf("%w: seul un échange accepté peut être terminé", apperr.ErrValidation)
	}
	m.credits[e.OwnerID] += credits
	e.Status = domain.StatusCompleted
	e.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	m.exchanges[exchangeID] = e
	return nil
}

func (m *Store) CancelExchange(ctx context.Context, exchangeID int, credits int, wasAccepted bool) error {
	e, ok := m.exchanges[exchangeID]
	if !ok {
		return fmt.Errorf("%w: échange %d", apperr.ErrNotFound, exchangeID)
	}
	if e.Status != domain.StatusPending && e.Status != domain.StatusAccepted {
		return fmt.Errorf("%w: cet échange ne peut plus être annulé", apperr.ErrValidation)
	}
	if wasAccepted {
		m.credits[e.RequesterID] += credits
	}
	e.Status = domain.StatusCancelled
	e.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	m.exchanges[exchangeID] = e
	return nil
}

func (m *Store) GetCreditBalance(ctx context.Context, userID int) (int, error) {
	if _, ok := m.users[userID]; !ok {
		return 0, fmt.Errorf("%w: utilisateur %d", apperr.ErrNotFound, userID)
	}
	return m.credits[userID], nil
}

func (m *Store) CreateReview(ctx context.Context, r domain.Review) (domain.Review, error) {
	for _, existing := range m.reviews {
		if existing.ExchangeID == r.ExchangeID && existing.AuthorID == r.AuthorID {
			return domain.Review{}, fmt.Errorf("%w: un avis a déjà été déposé pour cet échange", apperr.ErrValidation)
		}
	}
	m.nextReviewID++
	r.ID = m.nextReviewID
	r.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	m.reviews[r.ID] = r
	return r, nil
}

func (m *Store) ListReviewsByTarget(ctx context.Context, targetID int) ([]domain.Review, error) {
	out := []domain.Review{}
	for _, r := range m.reviews {
		if r.TargetID == targetID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *Store) ListReviewsByService(ctx context.Context, serviceID int) ([]domain.Review, error) {
	out := []domain.Review{}
	for _, r := range m.reviews {
		ex, ok := m.exchanges[r.ExchangeID]
		if !ok {
			continue
		}
		if ex.ServiceID == serviceID && r.TargetID == ex.OwnerID {
			out = append(out, r)
		}
	}
	return out, nil
}

// CreditBalance retourne le solde simulé d'un utilisateur (tests uniquement).
func (m *Store) CreditBalance(userID int) int {
	return m.credits[userID]
}

// SetCreditBalance force le solde d'un utilisateur (tests uniquement).
func (m *Store) SetCreditBalance(userID, balance int) {
	m.credits[userID] = balance
}

// SetStats injecte des statistiques prédéfinies (tests uniquement).
func (m *Store) SetStats(userID int, stats domain.UserStats) {
	m.stats[userID] = stats
}

// SeedReview ajoute un avis en mémoire (tests uniquement).
func (m *Store) SeedReview(r domain.Review) {
	if r.ID == 0 {
		m.nextReviewID++
		r.ID = m.nextReviewID
	}
	m.reviews[r.ID] = r
}

