package main

import (
	"context"
	"errors"
	"testing"
)

func TestReviewService_Create(t *testing.T) {
	store := newMockStore()
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusCompleted})
	svc := NewReviewService(store)
	ctx := context.Background()

	tests := []struct {
		name     string
		authorID int
		note     int
		wantErr  error
	}{
		{name: "success", authorID: 2, note: 5},
		{name: "invalid_note", authorID: 2, note: 6, wantErr: ErrValidation},
		{name: "forbidden", authorID: 3, note: 4, wantErr: ErrForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, tt.authorID, 1, CreateReviewInput{Note: tt.note})
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

func TestReviewService_NotCompletedExchange(t *testing.T) {
	store := newMockStore()
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusPending})
	svc := NewReviewService(store)

	_, err := svc.Create(context.Background(), 2, 1, CreateReviewInput{Note: 4})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestReviewService_DuplicateReview(t *testing.T) {
	store := newMockStore()
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusCompleted})
	svc := NewReviewService(store)
	ctx := context.Background()

	if _, err := svc.Create(ctx, 2, 1, CreateReviewInput{Note: 5}); err != nil {
		t.Fatalf("first review: %v", err)
	}
	_, err := svc.Create(ctx, 2, 1, CreateReviewInput{Note: 3})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestReviewService_TargetIsOwnerWhenRequesterReviews(t *testing.T) {
	store := newMockStore()
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusCompleted})
	svc := NewReviewService(store)

	r, err := svc.Create(context.Background(), 2, 1, CreateReviewInput{Note: 5, Commentaire: "top"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if r.TargetID != 1 {
		t.Fatalf("target_id = %d, want 1 (owner)", r.TargetID)
	}
}

func TestReviewService_ListByUserAndService(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusCompleted})
	store.reviews[1] = Review{ID: 1, ExchangeID: 1, AuthorID: 2, TargetID: 1, Note: 5}
	svc := NewReviewService(store)
	ctx := context.Background()

	reviews, err := svc.ListByUser(ctx, 1)
	if err != nil || len(reviews) != 1 {
		t.Fatalf("list by user: %v, len=%d", err, len(reviews))
	}
	svcReviews, err := svc.ListByService(ctx, 1)
	if err != nil || len(svcReviews) != 1 {
		t.Fatalf("list by service: %v, len=%d", err, len(svcReviews))
	}

	_, err = svc.ListByUser(ctx, 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
