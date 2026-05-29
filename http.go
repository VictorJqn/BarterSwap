package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, err error) {
	writeJSON(w, httpStatus(err), map[string]string{"error": err.Error()})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: JSON invalide : %v", ErrValidation, err)
	}
	return nil
}

func parseID(r *http.Request, name string) (int, error) {
	id, err := strconv.Atoi(r.PathValue(name))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%w: identifiant %q invalide", ErrValidation, name)
	}
	return id, nil
}
