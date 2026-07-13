package main

import (
	"context"
	"fmt"
	"strings"
)

// userRepository définit le contrat de persistance dont UserService a besoin.
// L'interface est déclarée ici (côté consommateur), pas dans le store.
type userRepository interface {
	CreateUser(ctx context.Context, pseudo, bio, ville string) (User, error)
	GetUserByID(ctx context.Context, id int) (User, error)
	UpdateUser(ctx context.Context, id int, pseudo, bio, ville string) (User, error)
	GetSkills(ctx context.Context, userID int) ([]Skill, error)
	ReplaceSkills(ctx context.Context, userID int, skills []Skill) error
}

// UserService contient la logique métier liée aux utilisateurs.
type UserService struct {
	repo userRepository
}

func NewUserService(repo userRepository) *UserService {
	return &UserService{repo: repo}
}

type CreateUserInput struct {
	Pseudo string
	Bio    string
	Ville  string
}

type UpdateUserInput struct {
	Pseudo string
	Bio    string
	Ville  string
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput) (User, error) {
	if strings.TrimSpace(in.Pseudo) == "" {
		return User{}, fmt.Errorf("%w: le pseudo est obligatoire", ErrValidation)
	}
	return s.repo.CreateUser(ctx, strings.TrimSpace(in.Pseudo), in.Bio, in.Ville)
}

func (s *UserService) GetByID(ctx context.Context, id int) (User, error) {
	if id <= 0 {
		return User{}, fmt.Errorf("%w: identifiant invalide", ErrValidation)
	}
	return s.repo.GetUserByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, actorID, targetID int, in UpdateUserInput) (User, error) {
	if actorID != targetID {
		return User{}, ErrForbidden
	}
	if strings.TrimSpace(in.Pseudo) == "" {
		return User{}, fmt.Errorf("%w: le pseudo est obligatoire", ErrValidation)
	}
	return s.repo.UpdateUser(ctx, targetID, strings.TrimSpace(in.Pseudo), in.Bio, in.Ville)
}

func (s *UserService) GetSkills(ctx context.Context, userID int) ([]Skill, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: identifiant invalide", ErrValidation)
	}
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.repo.GetSkills(ctx, userID)
}

func (s *UserService) ReplaceSkills(ctx context.Context, actorID, targetID int, skills []Skill) error {
	if actorID != targetID {
		return ErrForbidden
	}
	for _, sk := range skills {
		if strings.TrimSpace(sk.Nom) == "" {
			return fmt.Errorf("%w: le nom de compétence est obligatoire", ErrValidation)
		}
		if !validNiveau(sk.Niveau) {
			return fmt.Errorf("%w: niveau %q invalide", ErrValidation, sk.Niveau)
		}
	}
	return s.repo.ReplaceSkills(ctx, targetID, skills)
}

func (s *UserService) HasSkill(ctx context.Context, userID int, categorie string) (bool, error) {
	skills, err := s.repo.GetSkills(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, sk := range skills {
		if sk.Nom == categorie {
			return true, nil
		}
	}
	return false, nil
}
