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

func TestUserService_Create(t *testing.T) {
	store := mock.New()
	svc := service.NewUserService(store)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   service.CreateUserInput
		wantErr error
	}{
		{
			name:  "success",
			input: service.CreateUserInput{Pseudo: "alice", Ville: "Paris"},
		},
		{
			name:    "empty_pseudo",
			input:   service.CreateUserInput{Pseudo: "   "},
			wantErr: apperr.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := svc.Create(ctx, tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.Pseudo != tt.input.Pseudo {
				t.Fatalf("pseudo = %q, want %q", u.Pseudo, tt.input.Pseudo)
			}
			if u.CreditBalance != domain.WelcomeCredits {
				t.Fatalf("credit_balance = %d, want %d", u.CreditBalance, domain.WelcomeCredits)
			}
		})
	}
}

func TestUserService_GetByIDInvalid(t *testing.T) {
	svc := service.NewUserService(mock.New())
	_, err := svc.GetByID(context.Background(), 0)
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("expected apperr.ErrValidation, got %v", err)
	}
}

func TestUserService_UpdateForbidden(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	svc := service.NewUserService(store)

	_, err := svc.Update(context.Background(), 2, 1, service.UpdateUserInput{Pseudo: "bob"})
	if !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("expected apperr.ErrForbidden, got %v", err)
	}
}

func TestUserService_GetSkills(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})
	svc := service.NewUserService(store)

	skills, err := svc.GetSkills(context.Background(), 1)
	if err != nil || len(skills) != 1 {
		t.Fatalf("get skills: %v, len=%d", err, len(skills))
	}

	_, err = svc.GetSkills(context.Background(), 99)
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("expected apperr.ErrNotFound, got %v", err)
	}
}

func TestUserService_ReplaceSkillsValidation(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	svc := service.NewUserService(store)
	ctx := context.Background()

	tests := []struct {
		name    string
		skills  []domain.Skill
		wantErr error
	}{
		{
			name: "success",
			skills: []domain.Skill{
				{Nom: "Jardinage", Niveau: "expert"},
			},
		},
		{
			name:    "empty_skill_name",
			skills:  []domain.Skill{{Nom: " ", Niveau: "expert"}},
			wantErr: apperr.ErrValidation,
		},
		{
			name:    "invalid_level",
			skills:  []domain.Skill{{Nom: "Jardinage", Niveau: "pro"}},
			wantErr: apperr.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ReplaceSkills(ctx, 1, 1, tt.skills)
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

func TestUserService_ReplaceSkillsForbidden(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, nil)
	svc := service.NewUserService(store)

	err := svc.ReplaceSkills(context.Background(), 2, 1, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})
	if !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("expected apperr.ErrForbidden, got %v", err)
	}
}

func TestUserService_HasSkill(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 10, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})
	svc := service.NewUserService(store)
	ctx := context.Background()

	has, err := svc.HasSkill(ctx, 1, "Jardinage")
	if err != nil || !has {
		t.Fatalf("has skill: %v, %v", has, err)
	}
	has, err = svc.HasSkill(ctx, 1, "Cuisine")
	if err != nil || has {
		t.Fatalf("missing skill: %v, %v", has, err)
	}
}

func TestUserService_Stats(t *testing.T) {
	store := mock.New()
	store.SeedUser(domain.User{ID: 1, Pseudo: "alice"}, 12, nil)
	store.SetStats(1, domain.UserStats{
		UserID:            1,
		ServicesActifs:    2,
		EchangesCompletes: 3,
		NoteMoyenne:       4.5,
		NbAvis:            2,
		TotalGagne:        5,
		TotalDepense:      3,
	})
	svc := service.NewUserService(store)

	stats, err := svc.Stats(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.ServicesActifs != 2 || stats.EchangesCompletes != 3 || stats.CreditBalance != 12 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	_, err = svc.Stats(context.Background(), 99)
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("expected apperr.ErrNotFound, got %v", err)
	}
}
