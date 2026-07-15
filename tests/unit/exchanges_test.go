package unit

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"barterswap/internal/domain"
	"barterswap/internal/apperr"
	"barterswap/internal/service"
	"barterswap/tests/mock"
)

func TestExchangeService_Create(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.SeedUser(domain.User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.SeedService(domain.Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := service.NewExchangeService(store)
	ctx := context.Background()

	tests := []struct {
		name        int
		requesterID int
		serviceID   int
		wantErr     error
	}{
		{name: 1, requesterID: 2, serviceID: 1},
		{name: 2, requesterID: 1, serviceID: 1, wantErr: apperr.ErrValidation},
		{name: 3, requesterID: 2, serviceID: 99, wantErr: apperr.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("case_%d", tt.name), func(t *testing.T) {
			_, err := svc.Create(ctx, tt.requesterID, service.CreateExchangeInput{ServiceID: tt.serviceID})
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestExchangeService_InsufficientCredits(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.SeedUser(domain.User{ID: 2, Pseudo: "bob"}, 1, nil)
	store.SeedService(domain.Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := service.NewExchangeService(store)

	_, err := svc.Create(context.Background(), 2, service.CreateExchangeInput{ServiceID: 1})
	if !errors.Is(err, apperr.ErrInsufficientCredits) {
		t.Fatalf("expected apperr.ErrInsufficientCredits, got %v", err)
	}
}

func TestExchangeService_ConflictOnSecondPending(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.SeedUser(domain.User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.SeedUser(domain.User{ID: 3, Pseudo: "carol"}, 10, nil)
	store.SeedService(domain.Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := service.NewExchangeService(store)
	ctx := context.Background()

	if _, err := svc.Create(ctx, 2, service.CreateExchangeInput{ServiceID: 1}); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	_, err := svc.Create(ctx, 3, service.CreateExchangeInput{ServiceID: 1})
	if !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("expected apperr.ErrConflict, got %v", err)
	}
}

func TestExchangeService_CreditFlow(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.SeedUser(domain.User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.SeedService(domain.Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := service.NewExchangeService(store)
	ctx := context.Background()

	ex, err := svc.Create(ctx, 2, service.CreateExchangeInput{ServiceID: 1})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.Accept(ctx, 1, ex.ID); err != nil {
		t.Fatalf("accept: %v", err)
	}
	if store.CreditBalance(2) != 8 {
		t.Fatalf("bob balance after accept = %d, want 8", store.CreditBalance(2))
	}
	if _, err := svc.Complete(ctx, 1, ex.ID); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if store.CreditBalance(1) != 12 {
		t.Fatalf("alice balance after complete = %d, want 12", store.CreditBalance(1))
	}
}

func TestExchangeService_CancelRefundsCredits(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.SeedUser(domain.User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.SeedService(domain.Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := service.NewExchangeService(store)
	ctx := context.Background()

	ex, _ := svc.Create(ctx, 2, service.CreateExchangeInput{ServiceID: 1})
	_, _ = svc.Accept(ctx, 1, ex.ID)
	if _, err := svc.Cancel(ctx, 2, ex.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if store.CreditBalance(2) != 10 {
		t.Fatalf("bob balance after cancel = %d, want 10", store.CreditBalance(2))
	}
}

func TestExchangeService_GetByIDForbidden(t *testing.T) {
	store := mock.New()
	store.SeedExchange(domain.Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: domain.StatusPending})
	svc := service.NewExchangeService(store)

	_, err := svc.GetByID(context.Background(), 3, 1)
	if !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("expected apperr.ErrForbidden, got %v", err)
	}
}

func TestExchangeService_Reject(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.SeedUser(domain.User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.SeedService(domain.Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := service.NewExchangeService(store)
	ctx := context.Background()

	ex, _ := svc.Create(ctx, 2, service.CreateExchangeInput{ServiceID: 1})
	ex, err := svc.Reject(ctx, 1, ex.ID)
	if err != nil || ex.Status != domain.StatusRejected {
		t.Fatalf("reject: %v, status=%s", err, ex.Status)
	}
}

func TestExchangeService_ListInvalidStatus(t *testing.T) {
	svc := service.NewExchangeService(mock.New())
	_, err := svc.List(context.Background(), 1, "invalid")
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("expected apperr.ErrValidation, got %v", err)
	}
}

func TestExchangeService_ListByStatus(t *testing.T) {
	store := mock.New()
	store.SeedExchange(domain.Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: domain.StatusPending})
	store.SeedExchange(domain.Exchange{ID: 2, ServiceID: 2, RequesterID: 3, OwnerID: 1, Status: domain.StatusCompleted})
	svc := service.NewExchangeService(store)

	list, err := svc.List(context.Background(), 1, domain.StatusPending)
	if err != nil || len(list) != 1 || list[0].Status != domain.StatusPending {
		t.Fatalf("list: %v, %+v", err, list)
	}
}

func TestExchangeService_CreateInactiveService(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.SeedUser(domain.User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.SeedService(domain.Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: false})
	svc := service.NewExchangeService(store)

	_, err := svc.Create(context.Background(), 2, service.CreateExchangeInput{ServiceID: 1})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("expected apperr.ErrValidation, got %v", err)
	}
}

func TestExchangeService_CancelPendingNoRefund(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.SeedUser(domain.User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.SeedService(domain.Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := service.NewExchangeService(store)
	ctx := context.Background()

	ex, _ := svc.Create(ctx, 2, service.CreateExchangeInput{ServiceID: 1})
	if _, err := svc.Cancel(ctx, 2, ex.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if store.CreditBalance(2) != 10 {
		t.Fatalf("bob balance = %d, want 10 (no refund before accept)", store.CreditBalance(2))
	}
}

func TestExchangeService_AcceptForbidden(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.SeedUser(domain.User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.SeedService(domain.Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := service.NewExchangeService(store)
	ctx := context.Background()

	ex, _ := svc.Create(ctx, 2, service.CreateExchangeInput{ServiceID: 1})
	_, err := svc.Accept(ctx, 2, ex.ID)
	if !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("expected apperr.ErrForbidden, got %v", err)
	}
}
