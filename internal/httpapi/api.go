package httpapi

import (
	"net/http"

	"barterswap/internal/service"
)

// API regroupe les handlers HTTP (couche présentation).
// Chaque handler délègue la logique métier aux services.
type API struct {
	users     *service.UserService
	services  *service.ServiceService
	exchanges *service.ExchangeService
	reviews   *service.ReviewService
}

// NewAPI construit la couche HTTP en injectant les services métier.
func NewAPI(users *service.UserService, services *service.ServiceService, exchanges *service.ExchangeService, reviews *service.ReviewService) *API {
	return &API{users: users, services: services, exchanges: exchanges, reviews: reviews}
}

// Register attache toutes les routes REST de l'API au ServeMux fourni.
func (api *API) Register(mux *http.ServeMux) {
	// Utilisateurs
	mux.HandleFunc("POST /api/users", api.createUser)
	mux.HandleFunc("GET /api/users/{id}", api.getUser)
	mux.HandleFunc("PUT /api/users/{id}", api.updateUser)
	mux.HandleFunc("GET /api/users/{id}/skills", api.getSkills)
	mux.HandleFunc("GET /api/users/{id}/stats", api.getUserStats)
	mux.HandleFunc("PUT /api/users/{id}/skills", api.replaceSkills)

	// Services
	mux.HandleFunc("GET /api/services", api.listServices)
	mux.HandleFunc("POST /api/services", api.createService)
	mux.HandleFunc("GET /api/services/{id}", api.getService)
	mux.HandleFunc("PUT /api/services/{id}", api.updateService)
	mux.HandleFunc("DELETE /api/services/{id}", api.deleteService)

	// Échanges
	mux.HandleFunc("POST /api/exchanges", api.createExchange)
	mux.HandleFunc("GET /api/exchanges", api.listExchanges)
	mux.HandleFunc("GET /api/exchanges/{id}", api.getExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/accept", api.acceptExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/reject", api.rejectExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/complete", api.completeExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/cancel", api.cancelExchange)

	// Évaluations
	mux.HandleFunc("POST /api/exchanges/{id}/review", api.createReview)
	mux.HandleFunc("GET /api/users/{id}/reviews", api.listUserReviews)
	mux.HandleFunc("GET /api/services/{id}/reviews", api.listServiceReviews)
}
