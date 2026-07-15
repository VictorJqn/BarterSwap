package service

import (
	"barterswap/internal/apperr"
	"barterswap/internal/domain"
	"context"
	"fmt"
	"strings"
)

// UserRepository définit le contrat de persistance dont UserService a besoin.
// L'interface est déclarée ici (côté consommateur), pas dans le store.
type UserRepository interface {
	CreateUser(ctx context.Context, pseudo, bio, ville string) (domain.User, error)
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	UpdateUser(ctx context.Context, id int, pseudo, bio, ville string) (domain.User, error)
	GetSkills(ctx context.Context, userID int) ([]domain.Skill, error)
	ReplaceSkills(ctx context.Context, userID int, skills []domain.Skill) error
	GetUserStats(ctx context.Context, userID int) (domain.UserStats, error)
}

// UserService contient la logique métier liée aux utilisateurs.
type UserService struct {
	repo UserRepository
}

// NewUserService instancie le service métier des utilisateurs.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUserInput contient les champs pour l'inscription d'un nouvel utilisateur.
type CreateUserInput struct {
	Pseudo string
	Bio    string
	Ville  string
}

// UpdateUserInput contient les champs modifiables d'un profil utilisateur.
type UpdateUserInput struct {
	Pseudo string
	Bio    string
	Ville  string
}

// Create inscrit un utilisateur et lui attribue les crédits de bienvenue.
func (s *UserService) Create(ctx context.Context, in CreateUserInput) (domain.User, error) {
	if strings.TrimSpace(in.Pseudo) == "" {
		return domain.User{}, fmt.Errorf("%w: le pseudo est obligatoire", apperr.ErrValidation)
	}
	return s.repo.CreateUser(ctx, strings.TrimSpace(in.Pseudo), in.Bio, in.Ville)
}

// GetByID retourne le profil public d'un utilisateur, avec compétences et solde.
func (s *UserService) GetByID(ctx context.Context, id int) (domain.User, error) {
	if id <= 0 {
		return domain.User{}, fmt.Errorf("%w: identifiant invalide", apperr.ErrValidation)
	}
	return s.repo.GetUserByID(ctx, id)
}

// Update modifie le profil de l'utilisateur connecté (actorID doit égaler targetID).
func (s *UserService) Update(ctx context.Context, actorID, targetID int, in UpdateUserInput) (domain.User, error) {
	if actorID != targetID {
		return domain.User{}, apperr.ErrForbidden
	}
	if strings.TrimSpace(in.Pseudo) == "" {
		return domain.User{}, fmt.Errorf("%w: le pseudo est obligatoire", apperr.ErrValidation)
	}
	return s.repo.UpdateUser(ctx, targetID, strings.TrimSpace(in.Pseudo), in.Bio, in.Ville)
}

// GetSkills liste les compétences déclarées par un utilisateur.
func (s *UserService) GetSkills(ctx context.Context, userID int) ([]domain.Skill, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: identifiant invalide", apperr.ErrValidation)
	}
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		return nil, err
	}
	return s.repo.GetSkills(ctx, userID)
}

// ReplaceSkills remplace l'ensemble des compétences d'un utilisateur.
func (s *UserService) ReplaceSkills(ctx context.Context, actorID, targetID int, skills []domain.Skill) error {
	if actorID != targetID {
		return apperr.ErrForbidden
	}
	for _, sk := range skills {
		if strings.TrimSpace(sk.Nom) == "" {
			return fmt.Errorf("%w: le nom de compétence est obligatoire", apperr.ErrValidation)
		}
		if !domain.ValidNiveau(sk.Niveau) {
			return fmt.Errorf("%w: niveau %q invalide", apperr.ErrValidation, sk.Niveau)
		}
	}
	return s.repo.ReplaceSkills(ctx, targetID, skills)
}

// Stats retourne les indicateurs d'activité agrégés d'un utilisateur.
func (s *UserService) Stats(ctx context.Context, userID int) (domain.UserStats, error) {
	if _, err := s.repo.GetUserByID(ctx, userID); err != nil {
		return domain.UserStats{}, err
	}
	return s.repo.GetUserStats(ctx, userID)
}

// HasSkill indique si l'utilisateur possède la compétence correspondant à une catégorie.
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
