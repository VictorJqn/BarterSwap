package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"barterswap/internal/apperr"
)

// WriteJSON écrit une réponse JSON avec le code de statut donné.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	WriteJSON(w, status, v)
}

func writeError(w http.ResponseWriter, err error) {
	writeJSON(w, apperr.HTTPStatus(err), map[string]string{"error": err.Error()})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: JSON invalide : %v", apperr.ErrValidation, err)
	}
	return nil
}

func parseID(r *http.Request, name string) (int, error) {
	id, err := parseIDFromString(r.PathValue(name))
	if err != nil {
		return 0, fmt.Errorf("%w: identifiant %q invalide", apperr.ErrValidation, name)
	}
	return id, nil
}

func parseIDFromString(raw string) (int, error) {
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%w: identifiant invalide", apperr.ErrValidation)
	}
	return id, nil
}
