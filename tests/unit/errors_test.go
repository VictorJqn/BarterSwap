package unit

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"barterswap/internal/apperr"
)

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, http.StatusOK},
		{"validation", apperr.ErrValidation, http.StatusBadRequest},
		{"validation_wrapped", fmt.Errorf("champ: %w", apperr.ErrValidation), http.StatusBadRequest},
		{"credits", apperr.ErrInsufficientCredits, http.StatusBadRequest},
		{"unauthorized", apperr.ErrUnauthorized, http.StatusUnauthorized},
		{"forbidden", apperr.ErrForbidden, http.StatusForbidden},
		{"not_found", apperr.ErrNotFound, http.StatusNotFound},
		{"conflict", apperr.ErrConflict, http.StatusConflict},
		{"unknown", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := apperr.HTTPStatus(tt.err); got != tt.want {
				t.Fatalf("apperr.HTTPStatus() = %d, want %d", got, tt.want)
			}
		})
	}
}


