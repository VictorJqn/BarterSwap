package unit

import (
	"context"
	"errors"
	"testing"
	"barterswap/internal/domain"
	"barterswap/internal/apperr"
	"barterswap/internal/service"
	"barterswap/tests/mock"
)

func TestServiceService_Create(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})
	userSvc := service.NewUserService(store)
	svc := service.NewServiceService(store, userSvc)
	ctx := context.Background()

	valid := service.CreateServiceInput{
		Titre:        "Tonte",
		Categorie:    "Jardinage",
		DureeMinutes: 60,
		Credits:      2,
		Ville:        "Paris",
	}

	tests := []struct {
		name       string
		providerID int
		input      service.CreateServiceInput
		wantErr    error
	}{
		{name: "success", providerID: 1, input: valid},
		{
			name:       "missing_skill",
			providerID: 1,
			input: service.CreateServiceInput{
				Titre:        "Cours",
				Categorie:    "Cuisine",
				DureeMinutes: 60,
				Credits:      2,
			},
			wantErr: apperr.ErrValidation,
		},
		{
			name:       "invalid_category",
			providerID: 1,
			input: service.CreateServiceInput{
				Titre:        "Test",
				Categorie:    "Inconnu",
				DureeMinutes: 60,
				Credits:      2,
			},
			wantErr: apperr.ErrValidation,
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
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})
	store.SeedService(domain.Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	userSvc := service.NewUserService(store)
	svc := service.NewServiceService(store, userSvc)

	_, err := svc.Update(context.Background(), 2, 1, service.UpdateServiceInput{
		Titre:        "Hack",
		Categorie:    "Jardinage",
		DureeMinutes: 60,
		Credits:      2,
		Actif:        true,
	})
	if !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("expected apperr.ErrForbidden, got %v", err)
	}
}

func TestServiceService_ListInvalidCategory(t *testing.T) {
	store := mock.New()
	userSvc := service.NewUserService(store)
	svc := service.NewServiceService(store, userSvc)

	_, err := svc.List(context.Background(), service.ListServicesInput{Categorie: "Invalide"})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("expected apperr.ErrValidation, got %v", err)
	}
}

func TestServiceService_UpdateAndDelete(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})
	store.SeedService(domain.Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	userSvc := service.NewUserService(store)
	svc := service.NewServiceService(store, userSvc)
	ctx := context.Background()

	updated, err := svc.Update(ctx, 1, 1, service.UpdateServiceInput{
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
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("expected apperr.ErrNotFound after delete, got %v", err)
	}
}

func TestServiceService_ValidationErrors(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})
	userSvc := service.NewUserService(store)
	svc := service.NewServiceService(store, userSvc)

	_, err := svc.Create(context.Background(), 1, service.CreateServiceInput{
		Titre: "", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2,
	})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("expected apperr.ErrValidation for empty title, got %v", err)
	}

	_, err = svc.Create(context.Background(), 1, service.CreateServiceInput{
		Titre: "Test", Categorie: "Jardinage", DureeMinutes: 0, Credits: 2,
	})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("expected apperr.ErrValidation for invalid duration, got %v", err)
	}
}
