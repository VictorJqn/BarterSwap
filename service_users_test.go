package main

import (
	"context"
	"errors"
	"testing"
)

func TestUserService_Create(t *testing.T) {
	store := newMockStore()
	svc := NewUserService(store)
	ctx := context.Background()

	tests := []struct {
		name    string
		input   CreateUserInput
		wantErr error
	}{
		{
			name:  "success",
			input: CreateUserInput{Pseudo: "alice", Ville: "Paris"},
		},
		{
			name:    "empty_pseudo",
			input:   CreateUserInput{Pseudo: "   "},
			wantErr: ErrValidation,
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
			if u.CreditBalance != welcomeCredits {
				t.Fatalf("credit_balance = %d, want %d", u.CreditBalance, welcomeCredits)
			}
		})
	}
}

func TestUserService_GetByIDInvalid(t *testing.T) {
	svc := NewUserService(newMockStore())
	_, err := svc.GetByID(context.Background(), 0)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestUserService_UpdateForbidden(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	svc := NewUserService(store)

	_, err := svc.Update(context.Background(), 2, 1, UpdateUserInput{Pseudo: "bob"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUserService_GetSkills(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, []Skill{{Nom: "Jardinage", Niveau: "expert"}})
	svc := NewUserService(store)

	skills, err := svc.GetSkills(context.Background(), 1)
	if err != nil || len(skills) != 1 {
		t.Fatalf("get skills: %v, len=%d", err, len(skills))
	}

	_, err = svc.GetSkills(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserService_ReplaceSkillsValidation(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	svc := NewUserService(store)
	ctx := context.Background()

	tests := []struct {
		name    string
		skills  []Skill
		wantErr error
	}{
		{
			name: "success",
			skills: []Skill{
				{Nom: "Jardinage", Niveau: "expert"},
			},
		},
		{
			name:    "empty_skill_name",
			skills:  []Skill{{Nom: " ", Niveau: "expert"}},
			wantErr: ErrValidation,
		},
		{
			name:    "invalid_level",
			skills:  []Skill{{Nom: "Jardinage", Niveau: "pro"}},
			wantErr: ErrValidation,
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
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	svc := NewUserService(store)

	err := svc.ReplaceSkills(context.Background(), 2, 1, []Skill{{Nom: "Jardinage", Niveau: "expert"}})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUserService_HasSkill(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, []Skill{{Nom: "Jardinage", Niveau: "expert"}})
	svc := NewUserService(store)
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
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 12, nil)
	store.stats[1] = UserStats{
		UserID:            1,
		ServicesActifs:    2,
		EchangesCompletes: 3,
		NoteMoyenne:       4.5,
		NbAvis:            2,
		TotalGagne:        5,
		TotalDepense:      3,
	}
	svc := NewUserService(store)

	stats, err := svc.Stats(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.ServicesActifs != 2 || stats.EchangesCompletes != 3 || stats.CreditBalance != 12 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	_, err = svc.Stats(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
