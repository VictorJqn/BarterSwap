package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func apiRequest(t *testing.T, handler http.Handler, method, path string, body any, userID string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return out
}

func TestAPICreateUser(t *testing.T) {
	handler := newTestAPI(newMockStore())

	rec := apiRequest(t, handler, http.MethodPost, "/api/users", map[string]string{
		"pseudo": "alice",
		"ville":  "Paris",
	}, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	user := decodeBody[User](t, rec)
	if user.Pseudo != "alice" || user.CreditBalance != welcomeCredits {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestAPICreateUserEmptyPseudo(t *testing.T) {
	handler := newTestAPI(newMockStore())
	rec := apiRequest(t, handler, http.MethodPost, "/api/users", map[string]string{"pseudo": ""}, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAPIServiceWithoutSkill(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "bob"}, 10, nil)
	handler := newTestAPI(store)

	rec := apiRequest(t, handler, http.MethodPost, "/api/services", map[string]any{
		"titre":         "Cours cuisine",
		"categorie":     "Cuisine",
		"duree_minutes": 60,
		"credits":       2,
	}, "1")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAPIExchangeOwnService(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	handler := newTestAPI(store)

	rec := apiRequest(t, handler, http.MethodPost, "/api/exchanges", map[string]int{"service_id": 1}, "1")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAPIExchangeUnauthorized(t *testing.T) {
	handler := newTestAPI(newMockStore())
	rec := apiRequest(t, handler, http.MethodPost, "/api/exchanges", map[string]int{"service_id": 1}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestAPIExchangeConflict(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedUser(User{ID: 3, Pseudo: "carol"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	handler := newTestAPI(store)

	if apiRequest(t, handler, http.MethodPost, "/api/exchanges", map[string]int{"service_id": 1}, "2").Code != http.StatusCreated {
		t.Fatal("first exchange should succeed")
	}
	rec := apiRequest(t, handler, http.MethodPost, "/api/exchanges", map[string]int{"service_id": 1}, "3")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
}

func TestAPIFullExchangeFlow(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, []Skill{{Nom: "Jardinage", Niveau: "expert"}})
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	handler := newTestAPI(store)

	svcRec := apiRequest(t, handler, http.MethodPost, "/api/services", map[string]any{
		"titre":         "Tonte",
		"categorie":     "Jardinage",
		"duree_minutes": 60,
		"credits":       2,
		"ville":         "Paris",
	}, "1")
	if svcRec.Code != http.StatusCreated {
		t.Fatalf("create service: status %d", svcRec.Code)
	}
	service := decodeBody[Service](t, svcRec)

	exRec := apiRequest(t, handler, http.MethodPost, "/api/exchanges", map[string]int{"service_id": service.ID}, "2")
	if exRec.Code != http.StatusCreated {
		t.Fatalf("create exchange: status %d", exRec.Code)
	}
	ex := decodeBody[Exchange](t, exRec)

	acceptRec := apiRequest(t, handler, http.MethodPut, "/api/exchanges/"+strconv.Itoa(ex.ID)+"/accept", nil, "1")
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("accept: status %d", acceptRec.Code)
	}

	bobRec := apiRequest(t, handler, http.MethodGet, "/api/users/2", nil, "")
	bob := decodeBody[User](t, bobRec)
	if bob.CreditBalance != 8 {
		t.Fatalf("bob credits = %d, want 8", bob.CreditBalance)
	}

	completeRec := apiRequest(t, handler, http.MethodPut, "/api/exchanges/"+strconv.Itoa(ex.ID)+"/complete", nil, "1")
	if completeRec.Code != http.StatusOK {
		t.Fatalf("complete: status %d", completeRec.Code)
	}

	aliceRec := apiRequest(t, handler, http.MethodGet, "/api/users/1", nil, "")
	alice := decodeBody[User](t, aliceRec)
	if alice.CreditBalance != 12 {
		t.Fatalf("alice credits = %d, want 12", alice.CreditBalance)
	}
}

func TestAPIReviewOnPendingExchange(t *testing.T) {
	store := newMockStore()
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusPending})
	handler := newTestAPI(store)

	rec := apiRequest(t, handler, http.MethodPost, "/api/exchanges/1/review", map[string]any{
		"note": 5,
	}, "2")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAPIReviewDuplicate(t *testing.T) {
	store := newMockStore()
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusCompleted})
	handler := newTestAPI(store)

	body := map[string]any{"note": 5}
	if apiRequest(t, handler, http.MethodPost, "/api/exchanges/1/review", body, "2").Code != http.StatusCreated {
		t.Fatal("first review should succeed")
	}
	rec := apiRequest(t, handler, http.MethodPost, "/api/exchanges/1/review", body, "2")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAPIUserStats(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 12, nil)
	store.stats[1] = UserStats{
		UserID:            1,
		ServicesActifs:    1,
		EchangesCompletes: 2,
		TotalGagne:        4,
		TotalDepense:      2,
		NoteMoyenne:       4.5,
		NbAvis:            1,
	}
	handler := newTestAPI(store)

	rec := apiRequest(t, handler, http.MethodGet, "/api/users/1/stats", nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	stats := decodeBody[UserStats](t, rec)
	if stats.ServicesActifs != 1 || stats.EchangesCompletes != 2 || stats.CreditBalance != 12 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestAPIAdditionalEndpoints(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, []Skill{{Nom: "Jardinage", Niveau: "expert"}})
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true, Ville: "Paris"})
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusPending})
	store.reviews[1] = Review{ID: 1, ExchangeID: 1, AuthorID: 2, TargetID: 1, Note: 5}
	handler := newTestAPI(store)

	tests := []struct {
		name   string
		method string
		path   string
		body   any
		userID string
		want   int
	}{
		{"get user", http.MethodGet, "/api/users/1", nil, "", http.StatusOK},
		{"update user", http.MethodPut, "/api/users/1", map[string]string{"pseudo": "alice", "bio": "ok"}, "1", http.StatusOK},
		{"get skills", http.MethodGet, "/api/users/1/skills", nil, "", http.StatusOK},
		{"replace skills", http.MethodPut, "/api/users/1/skills", []Skill{{Nom: "Jardinage", Niveau: "expert"}}, "1", http.StatusOK},
		{"list services", http.MethodGet, "/api/services", nil, "", http.StatusOK},
		{"filter services", http.MethodGet, "/api/services?categorie=Jardinage&ville=Paris&search=tont", nil, "", http.StatusOK},
		{"get service", http.MethodGet, "/api/services/1", nil, "", http.StatusOK},
		{"update service", http.MethodPut, "/api/services/1", map[string]any{
			"titre": "Tonte+", "categorie": "Jardinage", "duree_minutes": 60, "credits": 2, "actif": true,
		}, "1", http.StatusOK},
		{"list exchanges", http.MethodGet, "/api/exchanges", nil, "1", http.StatusOK},
		{"get exchange", http.MethodGet, "/api/exchanges/1", nil, "1", http.StatusOK},
		{"reject exchange", http.MethodPut, "/api/exchanges/1/reject", nil, "1", http.StatusOK},
		{"list user reviews", http.MethodGet, "/api/users/1/reviews", nil, "", http.StatusOK},
		{"list service reviews", http.MethodGet, "/api/services/1/reviews", nil, "", http.StatusOK},
		{"invalid user id", http.MethodGet, "/api/users/abc", nil, "", http.StatusBadRequest},
		{"delete service", http.MethodDelete, "/api/services/1", nil, "1", http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := apiRequest(t, handler, tt.method, tt.path, tt.body, tt.userID)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d, body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestAPIInsufficientCredits(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 1, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	handler := newTestAPI(store)

	rec := apiRequest(t, handler, http.MethodPost, "/api/exchanges", map[string]int{"service_id": 1}, "2")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAPICancelExchange(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusAccepted})
	store.credits[2] = 8
	handler := newTestAPI(store)

	rec := apiRequest(t, handler, http.MethodPut, "/api/exchanges/1/cancel", nil, "2")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	ex := decodeBody[Exchange](t, rec)
	if ex.Status != StatusCancelled {
		t.Fatalf("status = %q, want cancelled", ex.Status)
	}
}

func TestAPIRejectExchange(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusPending})
	handler := newTestAPI(store)

	rec := apiRequest(t, handler, http.MethodPut, "/api/exchanges/1/reject", nil, "1")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestAPIListExchangesWithStatus(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusPending})
	store.seedExchange(Exchange{ID: 2, ServiceID: 2, RequesterID: 3, OwnerID: 1, Status: StatusCompleted})
	handler := newTestAPI(store)

	rec := apiRequest(t, handler, http.MethodGet, "/api/exchanges?status=pending", nil, "1")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	exchanges := decodeBody[[]Exchange](t, rec)
	if len(exchanges) != 1 || exchanges[0].Status != StatusPending {
		t.Fatalf("unexpected exchanges: %+v", exchanges)
	}
}

func TestAPIListExchangesInvalidStatus(t *testing.T) {
	handler := newTestAPI(newMockStore())
	rec := apiRequest(t, handler, http.MethodGet, "/api/exchanges?status=foo", nil, "1")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAPIInactiveServiceExchange(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, nil)
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: false})
	handler := newTestAPI(store)

	rec := apiRequest(t, handler, http.MethodPost, "/api/exchanges", map[string]int{"service_id": 1}, "2")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAPIForbiddenAndNotFound(t *testing.T) {
	store := newMockStore()
	store.seedUser(User{ID: 1, Pseudo: "alice"}, 10, []Skill{{Nom: "Jardinage", Niveau: "expert"}})
	store.seedUser(User{ID: 2, Pseudo: "bob"}, 10, nil)
	store.seedService(Service{ID: 1, ProviderID: 1, Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Actif: true})
	store.seedExchange(Exchange{ID: 1, ServiceID: 1, RequesterID: 2, OwnerID: 1, Status: StatusPending})
	handler := newTestAPI(store)

	tests := []struct {
		name   string
		method string
		path   string
		body   any
		userID string
		want   int
	}{
		{"get missing user", http.MethodGet, "/api/users/99", nil, "", http.StatusNotFound},
		{"update forbidden", http.MethodPut, "/api/users/1", map[string]string{"pseudo": "hack"}, "2", http.StatusForbidden},
		{"update service forbidden", http.MethodPut, "/api/services/1", map[string]any{
			"titre": "Hack", "categorie": "Jardinage", "duree_minutes": 60, "credits": 2, "actif": true,
		}, "2", http.StatusForbidden},
		{"delete service forbidden", http.MethodDelete, "/api/services/1", nil, "2", http.StatusForbidden},
		{"replace skills forbidden", http.MethodPut, "/api/users/1/skills", []Skill{{Nom: "Jardinage", Niveau: "expert"}}, "2", http.StatusForbidden},
		{"get exchange forbidden", http.MethodGet, "/api/exchanges/1", nil, "3", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := apiRequest(t, handler, tt.method, tt.path, tt.body, tt.userID)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d, body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestAPIInvalidJSONBody(t *testing.T) {
	handler := newTestAPI(newMockStore())
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestMiddlewareRecoveryAndCORS(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
	rec := httptest.NewRecorder()
	withRecovery(panicHandler).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("recovery status = %d", rec.Code)
	}

	corsRec := httptest.NewRecorder()
	withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(corsRec, httptest.NewRequest(http.MethodOptions, "/", nil))
	if corsRec.Code != http.StatusNoContent {
		t.Fatalf("options status = %d", corsRec.Code)
	}
}
