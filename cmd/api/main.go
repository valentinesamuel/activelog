package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/database"
	"github.com/valentinesamuel/activelog/internal/handlers"
	"github.com/valentinesamuel/activelog/internal/repository"
)

func main() {
	fmt.Println("🚒 Starting ActiveLog API...")

	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseUrl)
	if err != nil {
		log.Fatal("❌ 🛢️ Failed to connect to database \n", err)
	}
	defer db.Close()

	router := mux.NewRouter()

	activityRepo := repository.NewActivityRepository(db)

	healthHandler := handlers.NewHealthHandler()
	activityHandler := handlers.NewActivityHandler(activityRepo)

	router.Handle("/health", healthHandler).Methods("GET")

	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/activities", activityHandler.ListActivities).Methods("GET")
	api.HandleFunc("/activities", activityHandler.CreateActivity).Methods("POST")
	api.HandleFunc("/activities/{id}", activityHandler.GetActivity).Methods("POST")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "🪵 ActiveLog API v1", "version": "0.1.0"}`))
	}).Methods("GET")

	log.Printf("🚒 Server starting on :%s\n", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, router))

}
