package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	cfg := loadConfig()

	ctx := context.Background()
	db, err := openDB(ctx, cfg.databaseURL)
	if err != nil {
		log.Fatalf("base de données : %v", err)
	}
	defer db.Close()

	if err := migrate(ctx, db); err != nil {
		log.Fatalf("migration : %v", err)
	}
	log.Println("schéma vérifié / créé")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db indisponible"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	log.Printf("BarterSwap API à l'écoute sur :%s", cfg.port)
	if err := http.ListenAndServe(":"+cfg.port, mux); err != nil {
		log.Fatalf("serveur arrêté : %v", err)
	}
}
