package main

import (
	"errors"
	"net/http"
)

var (
	ErrValidation          = errors.New("données invalides")
	ErrUnauthorized        = errors.New("authentification requise")
	ErrForbidden           = errors.New("action interdite")
	ErrNotFound            = errors.New("ressource introuvable")
	ErrConflict            = errors.New("conflit")
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
