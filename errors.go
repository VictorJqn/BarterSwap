package main

import (
	"errors"
	"net/http"
)

// Erreurs sentinelles utilisées par la couche métier et traduites en codes HTTP.
var (
	// ErrValidation indique des données d'entrée invalides (400).
	ErrValidation = errors.New("données invalides")
	// ErrUnauthorized indique l'absence du header X-User-ID (401).
	ErrUnauthorized = errors.New("authentification requise")
	// ErrForbidden indique que l'utilisateur n'a pas le droit d'effectuer l'action (403).
	ErrForbidden = errors.New("action interdite")
	// ErrNotFound indique qu'une ressource demandée n'existe pas (404).
	ErrNotFound = errors.New("ressource introuvable")
	// ErrConflict indique un conflit métier, par ex. un service déjà réservé (409).
	ErrConflict = errors.New("conflit")
	// ErrInsufficientCredits indique un solde insuffisant pour l'opération (400).
	ErrInsufficientCredits = errors.New("crédits insuffisants")
)

// httpStatus traduit une erreur (même wrappée) en code de statut HTTP.
// Toute erreur non reconnue est considérée comme une erreur serveur (500).
func httpStatus(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, ErrValidation), errors.Is(err, ErrInsufficientCredits):
		return http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
