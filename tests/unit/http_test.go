package unit

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"barterswap/internal/apperr"
	"barterswap/internal/httpapi"
)

func TestWriteJSONEncodesPayload(t *testing.T) {
	rec := httptest.NewRecorder()
	httpapi.WriteJSON(rec, http.StatusOK, map[string]string{"status": "ok"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"ok"`)) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestHTTPStatusMapping(t *testing.T) {
	if apperr.HTTPStatus(apperr.ErrNotFound) != http.StatusNotFound {
		t.Fatal("expected 404")
	}
	if apperr.HTTPStatus(errors.New("unknown")) != http.StatusInternalServerError {
		t.Fatal("expected 500")
	}
}

func TestWithMiddleware(t *testing.T) {
	called := false
	handler := httpapi.WithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/users", nil))
	if !called || rec.Code != http.StatusOK {
		t.Fatalf("middleware failed: called=%v status=%d", called, rec.Code)
	}
}

func TestWithMiddlewareRecovery(t *testing.T) {
	handler := httpapi.WithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("recovery status = %d", rec.Code)
	}
}

func TestCORSOptions(t *testing.T) {
	handler := httpapi.WithMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("options status = %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header")
	}
}
