package main

import (
	"context"
	"log"
	"net/http"
	"rent-app/internal/config"
	"rent-app/internal/database"
	"time"

	"github.com/go-chi/chi/v5"
)

const timeout = 3 * time.Second

func main() {

	cfg, err := config.ConfigLoad()
	if err != nil {
		log.Fatal("failed to load config", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DSN)
	if err != nil {
		log.Fatal("failed connection to database", err)
	}
	defer db.Close()

	log.Println("successfully connected to database")
	mux := chi.NewRouter()

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome to the rent app!"))
	})

	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("database connection error"))
			return
		}
		w.Write([]byte("OK"))
	})

	log.Println("Starting server on port 8080...")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
