package main

import (
	"context"
	"fmt"
	"strings"
)

// serviceRepository définit le contrat de persistance dont ServiceService a besoin.
type serviceRepository interface {
	CreateService(ctx context.Context, s Service) (Service, error)
	GetServiceByID(ctx context.Context, id int) (Service, error)
	ListServices(ctx context.Context, filter ServiceFilter) ([]Service, error)
	UpdateService(ctx context.Context, s Service) (Service, error)
	DeleteService(ctx context.Context, id int) error
}

// ServiceFilter regroupe les critères de recherche côté base de données.
type ServiceFilter struct {
	Categorie string
	Ville     string
	Search    string
}

// ServiceService contient la logique métier liée aux annonces de services.
type ServiceService struct {
	repo  serviceRepository
	users *UserService
}

// NewServiceService instancie le service métier des annonces.
func NewServiceService(repo serviceRepository, users *UserService) *ServiceService {
	return &ServiceService{repo: repo, users: users}
}

// CreateServiceInput contient les champs pour publier une nouvelle annonce.
type CreateServiceInput struct {
	Titre        string
	Description  string
	Categorie    string
	DureeMinutes int
	Credits      int
	Ville        string
}

// UpdateServiceInput contient les champs modifiables d'une annonce existante.
type UpdateServiceInput struct {
	Titre        string
	Description  string
	Categorie    string
	DureeMinutes int
	Credits      int
	Ville        string
	Actif        bool
}

// ListServicesInput regroupe les filtres optionnels de recherche d'annonces.
type ListServicesInput struct {
	Categorie string
	Ville     string
	Search    string
}

// Create publie une annonce si le prestataire possède la compétence requise.
func (s *ServiceService) Create(ctx context.Context, providerID int, in CreateServiceInput) (Service, error) {
	if err := validateServiceFields(in.Titre, in.Categorie, in.DureeMinutes, in.Credits); err != nil {
		return Service{}, err
	}
	has, err := s.users.HasSkill(ctx, providerID, in.Categorie)
	if err != nil {
		return Service{}, err
	}
	if !has {
		return Service{}, fmt.Errorf("%w: compétence %q non possédée", ErrValidation, in.Categorie)
	}
	return s.repo.CreateService(ctx, Service{
		ProviderID:   providerID,
		Titre:        strings.TrimSpace(in.Titre),
		Description:  in.Description,
		Categorie:    in.Categorie,
		DureeMinutes: in.DureeMinutes,
		Credits:      in.Credits,
		Ville:        in.Ville,
		Actif:        true,
	})
}

// GetByID retourne le détail d'une annonce par son identifiant.
func (s *ServiceService) GetByID(ctx context.Context, id int) (Service, error) {
	if id <= 0 {
		return Service{}, fmt.Errorf("%w: identifiant invalide", ErrValidation)
	}
	return s.repo.GetServiceByID(ctx, id)
}

// List retourne les annonces filtrées côté serveur (catégorie, ville, recherche).
func (s *ServiceService) List(ctx context.Context, in ListServicesInput) ([]Service, error) {
	if in.Categorie != "" && !validCategorie(in.Categorie) {
		return nil, fmt.Errorf("%w: catégorie %q invalide", ErrValidation, in.Categorie)
	}
	return s.repo.ListServices(ctx, ServiceFilter{
		Categorie: in.Categorie,
		Ville:     in.Ville,
		Search:    in.Search,
	})
}

// Update modifie une annonce appartenant au prestataire connecté.
func (s *ServiceService) Update(ctx context.Context, actorID, serviceID int, in UpdateServiceInput) (Service, error) {
	if err := validateServiceFields(in.Titre, in.Categorie, in.DureeMinutes, in.Credits); err != nil {
		return Service{}, err
	}
	existing, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return Service{}, err
	}
	if existing.ProviderID != actorID {
		return Service{}, ErrForbidden
	}
	has, err := s.users.HasSkill(ctx, actorID, in.Categorie)
	if err != nil {
		return Service{}, err
	}
	if !has {
		return Service{}, fmt.Errorf("%w: compétence %q non possédée", ErrValidation, in.Categorie)
	}
	existing.Titre = strings.TrimSpace(in.Titre)
	existing.Description = in.Description
	existing.Categorie = in.Categorie
	existing.DureeMinutes = in.DureeMinutes
	existing.Credits = in.Credits
	existing.Ville = in.Ville
	existing.Actif = in.Actif
	return s.repo.UpdateService(ctx, existing)
}

// Delete supprime une annonce appartenant au prestataire connecté.
func (s *ServiceService) Delete(ctx context.Context, actorID, serviceID int) error {
	existing, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return err
	}
	if existing.ProviderID != actorID {
		return ErrForbidden
	}
	return s.repo.DeleteService(ctx, serviceID)
}

func validateServiceFields(titre, categorie string, duree, credits int) error {
	if strings.TrimSpace(titre) == "" {
		return fmt.Errorf("%w: le titre est obligatoire", ErrValidation)
	}
	if !validCategorie(categorie) {
		return fmt.Errorf("%w: catégorie %q invalide", ErrValidation, categorie)
	}
	if duree <= 0 {
		return fmt.Errorf("%w: la durée doit être positive", ErrValidation)
	}
	if credits <= 0 {
		return fmt.Errorf("%w: le coût en crédits doit être positif", ErrValidation)
	}
	return nil
}
