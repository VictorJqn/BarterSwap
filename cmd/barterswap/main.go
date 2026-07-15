package main

import (
	"context"
	"log"
	"net/http"

	"barterswap/internal/config"
	"barterswap/internal/database"
	"barterswap/internal/httpapi"
	"barterswap/internal/repository"
	"barterswap/internal/service"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("base de données : %v", err)
	}
	defer db.Close()

	if err := database.Migrate(ctx, db); err != nil {
		log.Fatalf("migration : %v", err)
	}
	log.Println("schéma vérifié / créé")

	store := repository.New(db)

	// Vérification à la compilation : Store satisfait les contrats des services.
	var (
		_ service.UserRepository     = store
		_ service.ServiceRepository  = store
		_ service.ExchangeRepository = store
		_ service.ReviewRepository   = store
	)

	userSvc := service.NewUserService(store)
	serviceSvc := service.NewServiceService(store, userSvc)
	exchangeSvc := service.NewExchangeService(store)
	reviewSvc := service.NewReviewService(store)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			httpapi.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db indisponible"})
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	api := httpapi.NewAPI(userSvc, serviceSvc, exchangeSvc, reviewSvc)
	api.Register(mux)

	log.Printf("BarterSwap API à l'écoute sur :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, httpapi.WithMiddleware(mux)); err != nil {
		log.Fatalf("serveur arrêté : %v", err)
	}
}
