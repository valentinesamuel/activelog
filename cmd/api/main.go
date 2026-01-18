package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/database"
	"github.com/valentinesamuel/activelog/internal/handlers"
	"github.com/valentinesamuel/activelog/internal/middleware"
	"github.com/valentinesamuel/activelog/internal/repository"
)

// Application holds all dependencies
type Application struct {
	DB              repository.DBConn
	DBCloser        interface{ Close() error } // For cleanup during shutdown
	Container       *container.Container       // DI container
	Broker          *broker.Broker             // Use case orchestrator
	HealthHandler   *handlers.HealthHandler
	ActivityHandler *handlers.ActivityHandler
	UserHandler     *handlers.UserHandler
	StatsHandler    *handlers.StatsHandler
	photoHandler    *handlers.ActivityPhotoHandler
}

func main() {
	fmt.Println("🚒 Starting ActiveLog API...")

	if err := run(); err != nil {
		log.Fatalf("❌ Application error: %v", err)
	}
}

// run orchestrates the application startup and shutdown
func run() error {
	// Load and validate configuration (loads .env file automatically)
	config.MustLoad()

	// Connect to database
	db, err := database.Connect(config.Database.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize application with dependencies
	app := &Application{
		DB:       db,
		DBCloser: db,
	}

	// Setup repositories and handlers
	app.setupDependencies()

	// Setup HTTP server
	server := app.newServer()

	// Run server with graceful shutdown
	return app.serve(server)
}

// setupDependencies initializes all repositories and handlers using DI container
// All dependencies are registered and resolved through the centralized container
func (app *Application) setupDependencies() {
	// Initialize container with all dependencies
	app.Container = setupContainer(app.DB)

	// Resolve core dependencies from container
	app.Broker = app.Container.MustResolve("broker").(*broker.Broker)

	// Resolve handlers from container
	app.HealthHandler = app.Container.MustResolve("healthHandler").(*handlers.HealthHandler)
	app.ActivityHandler = app.Container.MustResolve("activityHandler").(*handlers.ActivityHandler)
	app.UserHandler = app.Container.MustResolve("userHandler").(*handlers.UserHandler)
	app.StatsHandler = app.Container.MustResolve("statsHandler").(*handlers.StatsHandler)
	app.photoHandler = app.Container.MustResolve("activityPhotoHandler").(*handlers.ActivityPhotoHandler)
}

// setupRoutes configures all application routes and middleware
func (app *Application) setupRoutes() http.Handler {
	router := mux.NewRouter()

	// Global middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.CORS)
	router.Use(middleware.SecurityHeaders)

	// Health and root endpoints
	router.Handle("/health", app.HealthHandler).Methods("GET")
	router.HandleFunc("/", app.handleRoot).Methods("GET")

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Auth routes
	app.registerAuthRoutes(api)

	// Activity routes
	app.registerActivityRoutes(api)

	// Stats routes
	app.registerStatsRoutes(api)

	// User routes
	app.registerUserRoutes(api)

	return router
}

// handleRoot handles the root endpoint
func (app *Application) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "🪵 ActiveLog API v1", "version": "0.1.0"}`))
}

// registerAuthRoutes registers authentication routes
func (app *Application) registerAuthRoutes(router *mux.Router) {
	router.HandleFunc("/auth/register", app.UserHandler.CreateUser).Methods("POST")
	router.HandleFunc("/auth/login", app.UserHandler.LoginUser).Methods("POST")
}

// registerActivityRoutes registers activity CRUD routes
func (app *Application) registerActivityRoutes(router *mux.Router) {
	// router.Use(middleware.AuthMiddleware) // TODO: Enable when auth is ready
	router.HandleFunc("/activities", app.ActivityHandler.ListActivities).Methods("GET")
	router.HandleFunc("/activities", app.ActivityHandler.CreateActivity).Methods("POST")
	router.HandleFunc("/activities/stats", app.ActivityHandler.GetStats).Methods("GET")
	router.HandleFunc("/activities/{id}", app.ActivityHandler.GetActivity).Methods("GET")
	router.HandleFunc("/activities/{id}", app.ActivityHandler.UpdateActivity).Methods("PATCH")
	router.HandleFunc("/activities/{id}", app.ActivityHandler.DeleteActivity).Methods("DELETE")
	router.HandleFunc("/activities/{id}/photos", app.photoHandler.Upload).Methods("POST")
	router.HandleFunc("/activities/{id}/photos", app.photoHandler.GetActivityPhoto).Methods("GET")
}

// registerStatsRoutes registers statistics and analytics routes
func (app *Application) registerStatsRoutes(router *mux.Router) {
	// Create protected subrouter for stats endpoints
	statsRouter := router.PathPrefix("/stats").Subrouter()
	statsRouter.Use(middleware.AuthMiddleware)

	// Protected stats endpoints
	statsRouter.HandleFunc("/weekly", app.StatsHandler.GetWeeklyStats).Methods("GET")
	statsRouter.HandleFunc("/monthly", app.StatsHandler.GetMonthlyStats).Methods("GET")
	statsRouter.HandleFunc("/by-type", app.StatsHandler.GetActivityCountByType).Methods("GET")
}

// registerUserRoutes registers user-specific routes
func (app *Application) registerUserRoutes(router *mux.Router) {
	// Create protected subrouter for user endpoints
	userRouter := router.PathPrefix("/users/me").Subrouter()
	userRouter.Use(middleware.AuthMiddleware)

	// Protected user endpoints
	userRouter.HandleFunc("/summary", app.StatsHandler.GetUserActivitySummary).Methods("GET")
	userRouter.HandleFunc("/tags/top", app.StatsHandler.GetTopTags).Methods("GET")

	// Alternative user-scoped stats endpoints (as per Week 10 spec)
	userRouter.HandleFunc("/stats/weekly", app.StatsHandler.GetWeeklyStats).Methods("GET")
	userRouter.HandleFunc("/stats/monthly", app.StatsHandler.GetMonthlyStats).Methods("GET")
	userRouter.HandleFunc("/stats/by-type", app.StatsHandler.GetActivityCountByType).Methods("GET")
}

// newServer creates and configures the HTTP server
func (app *Application) newServer() *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Common.Port),
		Handler:      app.setupRoutes(),
		ReadTimeout:  45 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// serve starts the server and handles graceful shutdown
func (app *Application) serve(server *http.Server) error {
	// Create signal channel for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("🚒 Server starting on port %d...\n", config.Common.Port)
		serverErrors <- server.ListenAndServe()
	}()

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server failed to start: %w", err)
		}
	case sig := <-quit:
		log.Printf("🛑 Received signal: %v. Starting graceful shutdown...\n", sig)
		return app.gracefulShutdown(server)
	}

	return nil
}

// gracefulShutdown handles the graceful shutdown process
func (app *Application) gracefulShutdown(server *http.Server) error {
	// Create shutdown context with 30 second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown of HTTP server
	log.Println("⏳ Waiting for active connections to close...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
		// Force close if graceful shutdown fails
		if closeErr := server.Close(); closeErr != nil {
			log.Printf("❌ Error forcing server close: %v", closeErr)
		}
	} else {
		log.Println("✅ All connections closed gracefully")
	}

	// Close database connections
	log.Println("🔌 Closing database connections...")
	if err := app.DBCloser.Close(); err != nil {
		log.Printf("❌ Error closing database: %v", err)
		return err
	}
	log.Println("✅ Database connections closed")

	log.Println("👋 Server shutdown complete")
	return nil
}
