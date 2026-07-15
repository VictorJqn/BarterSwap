package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithMiddlewareAndLogging(t *testing.T) {
	called := false
	handler := withMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/users", nil))
	if !called {
		t.Fatal("handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestWithCORSAllowsGet(t *testing.T) {
	rec := httptest.NewRecorder()
	withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("missing CORS header")
	}
}

func TestRequireUserID(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		wantErr error
	}{
		{name: "missing", header: "", wantErr: ErrUnauthorized},
		{name: "invalid", header: "abc", wantErr: ErrValidation},
		{name: "valid", header: "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("X-User-ID", tt.header)
			}
			id, err := requireUserID(req)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil || id != 42 {
				t.Fatalf("id = %d, err = %v", id, err)
			}
		})
	}
}
