package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestExchangeService_Create(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := NewExchangeService(store)
	ctx := context.Background()

	tests := []struct {
		name        int
		requesterID int
		serviceID   int
		wantErr     error
	}{
		{name: 1, requesterID: 2, serviceID: 1},
		{name: 2, requesterID: 1, serviceID: 1, wantErr: ErrValidation},
		{name: 3, requesterID: 2, serviceID: 99, wantErr: ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("case_%d", tt.name), func(t *testing.T) {
			_, err := svc.Create(ctx, tt.requesterID, CreateExchangeInput{ServiceID: tt.serviceID})
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
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 1, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := NewExchangeService(store)

	_, err := svc.Create(context.Background(), 2, CreateExchangeInput{ServiceID: 1})
	if !errors.Is(err, ErrInsufficientCredits) {
		t.Fatalf("expected ErrInsufficientCredits, got %v", err)
	}
}

func TestExchangeService_ConflictOnSecondPending(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedUser(User{ID: 3, Pseudo: "carol"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := NewExchangeService(store)
	ctx := context.Background()

	if _, err := svc.Create(ctx, 2, CreateExchangeInput{ServiceID: 1}); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	_, err := svc.Create(ctx, 3, CreateExchangeInput{ServiceID: 1})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestExchangeService_CreditFlow(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := NewExchangeService(store)
	ctx := context.Background()

	ex, err := svc.Create(ctx, 2, CreateExchangeInput{ServiceID: 1})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.Accept(ctx, 1, ex.ID); err != nil {
		t.Fatalf("accept: %v", err)
	}
	if store.credits[2] != 8 {
		t.Fatalf("bob balance after accept = %d, want 8", store.credits[2])
	}
	if _, err := svc.Complete(ctx, 1, ex.ID); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if store.credits[1] != 12 {
		t.Fatalf("alice balance after complete = %d, want 12", store.credits[1])
	}
}

func TestExchangeService_CancelRefundsCredits(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := NewExchangeService(store)
	ctx := context.Background()

	ex, _ := svc.Create(ctx, 2, CreateExchangeInput{ServiceID: 1})
	_, _ = svc.Accept(ctx, 1, ex.ID)
	if _, err := svc.Cancel(ctx, 2, ex.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if store.credits[2] != 10 {
		t.Fatalf("bob balance after cancel = %d, want 10", store.credits[2])
	}
}

func TestExchangeService_GetByIDForbidden(t *testing.T) {
	store := newMockStore()
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusPending})
	svc := NewExchangeService(store)

	_, err := svc.GetByID(context.Background(), 3, 1)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestExchangeService_Reject(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := NewExchangeService(store)
	ctx := context.Background()

	ex, _ := svc.Create(ctx, 2, CreateExchangeInput{ServiceID: 1})
	ex, err := svc.Reject(ctx, 1, ex.ID)
	if err != nil || ex.Status != StatusRejected {
		t.Fatalf("reject: %v, status=%s", err, ex.Status)
	}
}

func TestExchangeService_ListInvalidStatus(t *testing.T) {
	svc := NewExchangeService(newMockStore())
	_, err := svc.List(context.Background(), 1, "invalid")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestExchangeService_ListByStatus(t *testing.T) {
	store := newMockStore()
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusPending})
	store.seedExchange(Exchange{ID: 2, ServiceID: 2, RequesterID: 3, OwnerID: 1, Status: StatusCompleted})
	svc := NewExchangeService(store)

	list, err := svc.List(context.Background(), 1, StatusPending)
	if err != nil || len(list) != 1 || list[0].Status != StatusPending {
		t.Fatalf("list: %v, %+v", err, list)
	}
}

func TestExchangeService_CreateInactiveService(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: false})
	svc := NewExchangeService(store)

	_, err := svc.Create(context.Background(), 2, CreateExchangeInput{ServiceID: 1})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestExchangeService_CancelPendingNoRefund(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := NewExchangeService(store)
	ctx := context.Background()

	ex, _ := svc.Create(ctx, 2, CreateExchangeInput{ServiceID: 1})
	if _, err := svc.Cancel(ctx, 2, ex.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if store.credits[2] != 10 {
		t.Fatalf("bob balance = %d, want 10 (no refund before accept)", store.credits[2])
	}
}

func TestExchangeService_AcceptForbidden(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	svc := NewExchangeService(store)
	ctx := context.Background()

	ex, _ := svc.Create(ctx, 2, CreateExchangeInput{ServiceID: 1})
	_, err := svc.Accept(ctx, 2, ex.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
