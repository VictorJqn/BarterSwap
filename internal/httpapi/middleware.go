package httpapi

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"barterswap/internal/apperr"
)

func WithMiddleware(h http.Handler) http.Handler {
	return withRecovery(withLogging(withCORS(h)))
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v\n%s", rec, debug.Stack())
				WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "erreur interne"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requireUserID(r *http.Request) (int, error) {
	raw := r.Header.Get("X-User-ID")
	if raw == "" {
		return 0, apperr.ErrUnauthorized
	}
	id, err := parseIDFromString(raw)
	if err != nil {
		return 0, err
	}
	return id, nil
}
