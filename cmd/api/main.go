package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/database"
	"github.com/valentinesamuel/activelog/internal/handlers"
	"github.com/valentinesamuel/activelog/internal/middleware"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/pkg/logger"
)

func main() {
	fmt.Println("🚒 Starting ActiveLog API...")

	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseUrl)
	if err != nil {
		log.Fatal("❌ 🛢️ Failed to connect to database \n", err)
	}
	defer db.Close()

	activityRepo := repository.NewActivityRepository(db)
	userRepo := repository.NewUserRepository(db)

	healthHandler := handlers.NewHealthHandler()
	activityHandler := handlers.NewActivityHandler(activityRepo)
	userHandler := handlers.NewUserHandler(userRepo)

	router := mux.NewRouter()

	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.CORS)
	router.Use(middleware.SecurityHeaders)

	router.Handle("/health", healthHandler).Methods("GET")

	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/activities", activityHandler.ListActivities).Methods("GET")
	api.HandleFunc("/activities", activityHandler.CreateActivity).Methods("POST")
	api.HandleFunc("/activities/stats", activityHandler.GetStats).Methods("GET")
	api.HandleFunc("/activities/{id}", activityHandler.GetActivity).Methods("GET")
	api.HandleFunc("/activities/{id}", activityHandler.UpdateActivity).Methods("PATCH")
	api.HandleFunc("/activities/{id}", activityHandler.DeleteActivity).Methods("DELETE")
	api.HandleFunc("/auth/register", userHandler.CreateUser).Methods("POST")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "🪵 ActiveLog API v1", "version": "0.1.0"}`))
	}).Methods("GET")

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  45 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info().Str("port", cfg.ServerPort).Msg("🚒 Server starting ...")
	log.Fatal(server.ListenAndServe())

}
