package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// mockStore simule la couche SQL en mémoire pour les tests unitaires et API.
type mockStore struct {
	users     map[int]User
	skills    map[int][]Skill
	services  map[int]Service
	exchanges map[int]Exchange
	reviews   map[int]Review
	credits   map[int]int
	stats     map[int]UserStats

	nextUserID     int
	nextServiceID  int
	nextExchangeID int
	nextReviewID   int
}

func newMockStore() *mockStore {
	return &mockStore{
		users:     make(map[int]User),
		skills:    make(map[int][]Skill),
		services:  make(map[int]Service),
		exchanges: make(map[int]Exchange),
		reviews:   make(map[int]Review),
		credits:   make(map[int]int),
		stats:     make(map[int]UserStats),
	}
}

func (m *mockStore) seedUser(u User, balance int, skills []Skill) {
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

func (m *mockStore) seedService(s Service) {
	if s.ID == 0 {
		m.nextServiceID++
		s.ID = m.nextServiceID
	}
	s.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	m.services[s.ID] = s
}

func (m *mockStore) seedExchange(e Exchange) {
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

func (m *mockStore) hasActiveExchange(serviceID int) bool {
	for _, ex := range m.exchanges {
		if ex.ServiceID == serviceID && (ex.Status == StatusPending || ex.Status == StatusAccepted) {
			return true
		}
	}
	return false
}

func (m *mockStore) CreateUser(ctx context.Context, pseudo, bio, ville string) (User, error) {
	m.nextUserID++
	u := User{
		ID:            m.nextUserID,
		Pseudo:        pseudo,
		Bio:           bio,
		Ville:         ville,
		CreditBalance: welcomeCredits,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		Skills:        []Skill{},
	}
	m.users[u.ID] = u
	m.credits[u.ID] = welcomeCredits
	m.skills[u.ID] = []Skill{}
	return u, nil
}

func (m *mockStore) GetUserByID(ctx context.Context, id int) (User, error) {
	u, ok := m.users[id]
	if !ok {
		return User{}, fmt.Errorf("%w: utilisateur %d", ErrNotFound, id)
	}
	u.CreditBalance = m.credits[id]
	u.Skills = m.skills[id]
	return u, nil
}

func (m *mockStore) UpdateUser(ctx context.Context, id int, pseudo, bio, ville string) (User, error) {
	u, ok := m.users[id]
	if !ok {
		return User{}, fmt.Errorf("%w: utilisateur %d", ErrNotFound, id)
	}
	u.Pseudo = pseudo
	u.Bio = bio
	u.Ville = ville
	m.users[id] = u
	return m.GetUserByID(ctx, id)
}

func (m *mockStore) GetSkills(ctx context.Context, userID int) ([]Skill, error) {
	if _, ok := m.users[userID]; !ok {
		return nil, fmt.Errorf("%w: utilisateur %d", ErrNotFound, userID)
	}
	return m.skills[userID], nil
}

func (m *mockStore) ReplaceSkills(ctx context.Context, userID int, skills []Skill) error {
	if _, ok := m.users[userID]; !ok {
		return fmt.Errorf("%w: utilisateur %d", ErrNotFound, userID)
	}
	m.skills[userID] = skills
	return nil
}

func (m *mockStore) GetUserStats(ctx context.Context, userID int) (UserStats, error) {
	if s, ok := m.stats[userID]; ok {
		s.CreditBalance = m.credits[userID]
		return s, nil
	}
	if _, ok := m.users[userID]; !ok {
		return UserStats{}, fmt.Errorf("%w: utilisateur %d", ErrNotFound, userID)
	}
	return UserStats{
		UserID:        userID,
		CreditBalance: m.credits[userID],
	}, nil
}

func (m *mockStore) CreateService(ctx context.Context, s Service) (Service, error) {
	m.nextServiceID++
	s.ID = m.nextServiceID
	s.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	m.services[s.ID] = s
	return s, nil
}

func (m *mockStore) GetServiceByID(ctx context.Context, id int) (Service, error) {
	s, ok := m.services[id]
	if !ok {
		return Service{}, fmt.Errorf("%w: service %d", ErrNotFound, id)
	}
	return s, nil
}

func (m *mockStore) ListServices(ctx context.Context, filter ServiceFilter) ([]Service, error) {
	out := []Service{}
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

func (m *mockStore) UpdateService(ctx context.Context, s Service) (Service, error) {
	if _, ok := m.services[s.ID]; !ok {
		return Service{}, fmt.Errorf("%w: service %d", ErrNotFound, s.ID)
	}
	m.services[s.ID] = s
	return s, nil
}

func (m *mockStore) DeleteService(ctx context.Context, id int) error {
	if _, ok := m.services[id]; !ok {
		return fmt.Errorf("%w: service %d", ErrNotFound, id)
	}
	delete(m.services, id)
	return nil
}

func (m *mockStore) CreateExchange(ctx context.Context, e Exchange) (Exchange, error) {
	if m.hasActiveExchange(e.ServiceID) {
		return Exchange{}, fmt.Errorf("%w: service déjà réservé", ErrConflict)
	}
	m.nextExchangeID++
	e.ID = m.nextExchangeID
	e.Status = StatusPending
	now := time.Now().UTC().Format(time.RFC3339)
	e.CreatedAt = now
	e.UpdatedAt = now
	m.exchanges[e.ID] = e
	return e, nil
}

func (m *mockStore) GetExchangeByID(ctx context.Context, id int) (Exchange, error) {
	e, ok := m.exchanges[id]
	if !ok {
		return Exchange{}, fmt.Errorf("%w: échange %d", ErrNotFound, id)
	}
	return e, nil
}

func (m *mockStore) ListExchanges(ctx context.Context, userID int, status string) ([]Exchange, error) {
	out := []Exchange{}
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

func (m *mockStore) AcceptExchange(ctx context.Context, exchangeID int, credits int) error {
	e, ok := m.exchanges[exchangeID]
	if !ok {
		return fmt.Errorf("%w: échange %d", ErrNotFound, exchangeID)
	}
	if e.Status != StatusPending {
		return fmt.Errorf("%w: seul un échange en attente peut être accepté", ErrValidation)
	}
	if m.credits[e.RequesterID] < credits {
		return ErrInsufficientCredits
	}
	m.credits[e.RequesterID] -= credits
	e.Status = StatusAccepted
	e.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	m.exchanges[exchangeID] = e
	return nil
}

func (m *mockStore) RejectExchange(ctx context.Context, exchangeID int) error {
	e, ok := m.exchanges[exchangeID]
	if !ok {
		return fmt.Errorf("%w: échange %d", ErrNotFound, exchangeID)
	}
	if e.Status != StatusPending {
		return fmt.Errorf("%w: seul un échange en attente peut être refusé", ErrValidation)
	}
	e.Status = StatusRejected
	e.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	m.exchanges[exchangeID] = e
	return nil
}

func (m *mockStore) CompleteExchange(ctx context.Context, exchangeID int, credits int) error {
	e, ok := m.exchanges[exchangeID]
	if !ok {
		return fmt.Errorf("%w: échange %d", ErrNotFound, exchangeID)
	}
	if e.Status != StatusAccepted {
		return fmt.Errorf("%w: seul un échange accepté peut être terminé", ErrValidation)
	}
	m.credits[e.OwnerID] += credits
	e.Status = StatusCompleted
	e.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	m.exchanges[exchangeID] = e
	return nil
}

func (m *mockStore) CancelExchange(ctx context.Context, exchangeID int, credits int, wasAccepted bool) error {
	e, ok := m.exchanges[exchangeID]
	if !ok {
		return fmt.Errorf("%w: échange %d", ErrNotFound, exchangeID)
	}
	if e.Status != StatusPending && e.Status != StatusAccepted {
		return fmt.Errorf("%w: cet échange ne peut plus être annulé", ErrValidation)
	}
	if wasAccepted {
		m.credits[e.RequesterID] += credits
	}
	e.Status = StatusCancelled
	e.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	m.exchanges[exchangeID] = e
	return nil
}

func (m *mockStore) GetCreditBalance(ctx context.Context, userID int) (int, error) {
	if _, ok := m.users[userID]; !ok {
		return 0, fmt.Errorf("%w: utilisateur %d", ErrNotFound, userID)
	}
	return m.credits[userID], nil
}

func (m *mockStore) CreateReview(ctx context.Context, r Review) (Review, error) {
	for _, existing := range m.reviews {
		if existing.ExchangeID == r.ExchangeID && existing.AuthorID == r.AuthorID {
			return Review{}, fmt.Errorf("%w: un avis a déjà été déposé pour cet échange", ErrValidation)
		}
	}
	m.nextReviewID++
	r.ID = m.nextReviewID
	r.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	m.reviews[r.ID] = r
	return r, nil
}

func (m *mockStore) ListReviewsByTarget(ctx context.Context, targetID int) ([]Review, error) {
	out := []Review{}
	for _, r := range m.reviews {
		if r.TargetID == targetID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *mockStore) ListReviewsByService(ctx context.Context, serviceID int) ([]Review, error) {
	out := []Review{}
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

func newTestAPI(store *mockStore) http.Handler {
	userSvc := NewUserService(store)
	serviceSvc := NewServiceService(store, userSvc)
	exchangeSvc := NewExchangeService(store)
	reviewSvc := NewReviewService(store)
	mux := http.NewServeMux()
	NewAPI(userSvc, serviceSvc, exchangeSvc, reviewSvc).Register(mux)
	return mux
}
