package api

import (
	"net/http"

	"barterswap/internal/httpapi"
	"barterswap/internal/service"
	"barterswap/tests/mock"
)

func newTestAPI(store *mock.Store) http.Handler {
	userSvc := service.NewUserService(store)
	serviceSvc := service.NewServiceService(store, userSvc)
	exchangeSvc := service.NewExchangeService(store)
	reviewSvc := service.NewReviewService(store)
	mux := http.NewServeMux()
	httpapi.NewAPI(userSvc, serviceSvc, exchangeSvc, reviewSvc).Register(mux)
	return mux
}
