package main

import (
	"context"
	"errors"
	"testing"
)

func TestServiceService_Create(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, []Skill{{Nom: "Jardinage", Niveau: "expert"}})
	userSvc := NewUserService(store)
	svc := NewServiceService(store, userSvc)
	ctx := context.Background()

	valid := CreateServiceInput{
		Titre:        "Tonte",
		Categorie:    "Jardinage",
		DureeMinutes: 60,
		Credits:      2,
		Ville:        "Paris",
	}

	tests := []struct {
		name       string
		providerID int
		input      CreateServiceInput
		wantErr    error
	}{
		{name: "success", providerID: 1, input: valid},
		{
			name:       "missing_skill",
			providerID: 1,
			input: CreateServiceInput{
				Titre:        "Cours",
				Categorie:    "Cuisine",
				DureeMinutes: 60,
				Credits:      2,
			},
			wantErr: ErrValidation,
		},
		{
			name:       "invalid_category",
			providerID: 1,
			input: CreateServiceInput{
				Titre:        "Test",
				Categorie:    "Inconnu",
				DureeMinutes: 60,
				Credits:      2,
			},
			wantErr: ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, tt.providerID, tt.input)
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

func TestServiceService_UpdateForbidden(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, []Skill{{Nom: "Jardinage", Niveau: "expert"}})
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	userSvc := NewUserService(store)
	svc := NewServiceService(store, userSvc)

	_, err := svc.Update(context.Background(), 2, 1, UpdateServiceInput{
		Titre:        "Hack",
		Categorie:    "Jardinage",
		DureeMinutes: 60,
		Credits:      2,
		Actif:        true,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestServiceService_ListInvalidCategory(t *testing.T) {
	store := newMockStore()
	userSvc := NewUserService(store)
	svc := NewServiceService(store, userSvc)

	_, err := svc.List(context.Background(), ListServicesInput{Categorie: "Invalide"})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestServiceService_UpdateAndDelete(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, []Skill{{Nom: "Jardinage", Niveau: "expert"}})
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	userSvc := NewUserService(store)
	svc := NewServiceService(store, userSvc)
	ctx := context.Background()

	updated, err := svc.Update(ctx, 1, 1, UpdateServiceInput{
		Titre:        "Tonte premium",
		Categorie:    "Jardinage",
		DureeMinutes: 90,
		Credits:      3,
		Ville:        "Paris",
		Actif:        false,
	})
	if err != nil || updated.Titre != "Tonte premium" || updated.Actif {
		t.Fatalf("update: %v, %+v", err, updated)
	}

	if err := svc.Delete(ctx, 1, 1); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = svc.GetByID(ctx, 1)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestServiceService_ValidationErrors(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, []Skill{{Nom: "Jardinage", Niveau: "expert"}})
	userSvc := NewUserService(store)
	svc := NewServiceService(store, userSvc)

	_, err := svc.Create(context.Background(), 1, CreateServiceInput{
		Titre: "", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2,
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation for empty title, got %v", err)
	}

	_, err = svc.Create(context.Background(), 1, CreateServiceInput{
		Titre: "Test", Categorie: "Jardinage", DureeMinutes: 0, Credits: 2,
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation for invalid duration, got %v", err)
	}
}
