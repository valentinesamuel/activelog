package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valentinesamuel/activelog/pkg/database"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/valentinesamuel/activelog/docs"
	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/application/broker/di"
	cacheDI "github.com/valentinesamuel/activelog/internal/adapters/cache/di"
	cacheTypes "github.com/valentinesamuel/activelog/internal/adapters/cache/types"
	"github.com/valentinesamuel/activelog/internal/platform/config"
	"github.com/valentinesamuel/activelog/internal/platform/container"
	"github.com/valentinesamuel/activelog/internal/platform/featureflags"
	"github.com/valentinesamuel/activelog/internal/handlers"
	handlerDI "github.com/valentinesamuel/activelog/internal/handlers/di"
	"github.com/valentinesamuel/activelog/internal/middleware"
	queueDI "github.com/valentinesamuel/activelog/internal/adapters/queue/di"
	queueTypes "github.com/valentinesamuel/activelog/internal/adapters/queue/types"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/platform/scheduler"
	schedulerDI "github.com/valentinesamuel/activelog/internal/platform/scheduler/di"
	"github.com/valentinesamuel/activelog/internal/adapters/webhook"
	webhookDI "github.com/valentinesamuel/activelog/internal/adapters/webhook/di"
	webhookTypes "github.com/valentinesamuel/activelog/internal/adapters/webhook/types"
	appwebsocket "github.com/valentinesamuel/activelog/internal/adapters/websocket"
)

// @title ActiveLog API
// @version 1.0
// @description Activity tracking REST API for logging and analyzing physical activities
// @host localhost:8080
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format: Bearer {token}

// Application holds all dependencies
type Application struct {
	DB              repository.DBConn
	DBCloser        interface{ Close() error } // For cleanup during shutdown
	Container       *container.Container       // DI container
	Broker          *broker.Broker             // Use case orchestrator
	Scheduler       *scheduler.Scheduler       // Cron scheduler
	RateLimiter     *middleware.RateLimiter    // Rate limiting middleware
	Flags           *featureflags.FeatureFlags
	FlagMiddleware  *featureflags.Middleware
	WSHub           *appwebsocket.Hub
	WSHandler       *appwebsocket.Handler
	HealthHandler   *handlers.HealthHandler
	ActivityHandler *handlers.ActivityHandler
	UserHandler     *handlers.UserHandler
	StatsHandler    *handlers.StatsHandler
	photoHandler    *handlers.ActivityPhotoHandler
	ExportHandler    *handlers.ExportHandler
	FeaturesHandler  *handlers.FeaturesHandler
	WebhookHandler   *handlers.WebhookHandler
	WebhookBus          webhookTypes.WebhookBusProvider
	WebhookDelivery     *webhook.Delivery
	WebhookRetryWorker  *webhook.RetryWorker
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

	//redis, err := cache.Connect()
	//if err != nil {
	//	return fmt.Errorf("failed to connect to Redis: %w", err)
	//}

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

// cacheRateLimitConfig writes the in-memory rate limit config to Redis at
// startup with a 48-hour TTL, then reads it back to verify it was stored.
func (app *Application) cacheRateLimitConfig(adapter cacheTypes.CacheAdapter) {
	ctx := context.Background()
	opts := cacheTypes.CacheOptions{
		DB:           cacheTypes.CacheDBRateLimits,
		PartitionKey: cacheTypes.CachePartitionRateLimitConfig,
	}

	cached := struct {
		CachedAt time.Time              `json:"cached_at"`
		Config   *config.RateLimitConfig `json:"config"`
	}{
		CachedAt: time.Now(),
		Config:   config.RateLimit,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		log.Printf("Warning: Failed to marshal rate limit config: %v", err)
		return
	}

	if err := adapter.Set(ctx, "config", string(data), 48*time.Hour, opts); err != nil {
		log.Printf("Warning: Failed to cache rate limit config to Redis: %v", err)
		return
	}
	log.Printf("Rate limit config cached to Redis (DB %d)", config.Cache.DBs.RateLimits)

	// Verify by reading back
	if val, err := adapter.Get(ctx, "config", opts); err == nil && val != "" {
		log.Printf("Rate limit config verified from Redis")
	} else {
		log.Printf("Warning: Rate limit config not found in Redis after write, using in-memory fallback")
	}
}

// setupDependencies initializes all repositories and handlers using DI container
// All dependencies are registered and resolved through the centralized container
func (app *Application) setupDependencies() {
	// Load feature flags
	app.Flags = featureflags.Load()
	app.FlagMiddleware = featureflags.NewMiddleware(app.Flags)
	app.FeaturesHandler = handlers.NewFeaturesHandler(app.Flags)

	// Create WebSocket hub
	app.WSHub = appwebsocket.NewHub()
	app.WSHandler = appwebsocket.NewHandler(app.WSHub)

	// Initialize container with all dependencies
	app.Container = setupContainer(app.DB, app.WSHub)

	// Resolve core dependencies from container
	app.Broker = app.Container.MustResolve(di.BrokerKey).(*broker.Broker)

	// Setup rate limiter using the multi-DB cache adapter
	resolvedAdapter := app.Container.MustResolve(cacheDI.CacheAdapterKey)
	cacheAdapter := resolvedAdapter.(cacheTypes.CacheAdapter)
	rlCacheProvider := resolvedAdapter.(cacheTypes.RateLimitCacheProvider)
	queueProvider := app.Container.MustResolve(queueDI.QueueProviderKey).(queueTypes.QueueProvider)

	// Write rate limit config to Redis at startup with 48h TTL
	app.cacheRateLimitConfig(cacheAdapter)

	app.RateLimiter = middleware.NewRateLimiter(rlCacheProvider, cacheAdapter, queueProvider, config.RateLimit)

	// Resolve scheduler from container
	app.Scheduler = app.Container.MustResolve(schedulerDI.SchedulerKey).(*scheduler.Scheduler)

	// Resolve handlers from container
	app.HealthHandler = app.Container.MustResolve(handlerDI.HealthHandlerKey).(*handlers.HealthHandler)
	app.ActivityHandler = app.Container.MustResolve(handlerDI.ActivityHandlerKey).(*handlers.ActivityHandler)
	app.UserHandler = app.Container.MustResolve(handlerDI.UserHandlerKey).(*handlers.UserHandler)
	app.StatsHandler = app.Container.MustResolve(handlerDI.StatsHandlerKey).(*handlers.StatsHandler)
	app.photoHandler = app.Container.MustResolve(handlerDI.ActivityPhotoHandlerKey).(*handlers.ActivityPhotoHandler)
	app.ExportHandler = app.Container.MustResolve(handlerDI.ExportHandlerKey).(*handlers.ExportHandler)
	app.WebhookHandler = app.Container.MustResolve(handlerDI.WebhookHandlerKey).(*handlers.WebhookHandler)

	// Resolve webhook bus, delivery, and retry worker from container
	app.WebhookDelivery = app.Container.MustResolve(webhookDI.WebhookDeliveryKey).(*webhook.Delivery)
	app.WebhookRetryWorker = app.Container.MustResolve(webhookDI.RetryWorkerKey).(*webhook.RetryWorker)
	app.WebhookBus = app.Container.MustResolve(webhookDI.WebhookBusKey).(webhookTypes.WebhookBusProvider)
}

// setupRoutes configures all application routes and middleware
func (app *Application) setupRoutes() http.Handler {
	router := mux.NewRouter()

	// Global middleware
	router.Use(middleware.TimingMiddleware)
	router.Use(middleware.MetricsMiddleware)
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.CORS)
	router.Use(middleware.SecurityHeaders)
	router.Use(app.RateLimiter.Middleware)

	// Health and root endpoints
	router.Handle("/health", app.HealthHandler).Methods("GET")
	router.HandleFunc("/", app.handleRoot).Methods("GET")

	router.Handle("/metrics", promhttp.Handler())

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Auth routes (public - no auth required)
	app.registerAuthRoutes(api)

	// Activity routes (protected)
	app.registerActivityRoutes(api)

	// Stats routes
	app.registerStatsRoutes(api)

	// User routes
	app.registerUserRoutes(api)

	// Export routes
	app.registerExportRoutes(api)

	// Features route
	app.registerFeaturesRoutes(api)

	// Webhook routes
	app.registerWebhookRoutes(api)

	// WebSocket route (protected - JWT via query param or header)
	wsRouter := router.PathPrefix("/ws").Subrouter()
	wsRouter.Use(middleware.AuthMiddleware)
	wsRouter.HandleFunc("", app.WSHandler.ServeWS)

	return router
}

// handleRoot handles the root endpoint
func (app *Application) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "🪵 ActiveLog API v1", "version": "0.1.0"}`))
}

// registerAuthRoutes registers authentication routes
func (app *Application) registerAuthRoutes(router *mux.Router) {
	authRouter := router.PathPrefix("/auth").Subrouter()

	authRouter.HandleFunc("/register", app.UserHandler.CreateUser).Methods("POST")
	authRouter.HandleFunc("/login", app.UserHandler.LoginUser).Methods("POST")
}

// registerActivityRoutes registers activity CRUD routes
func (app *Application) registerActivityRoutes(router *mux.Router) {
	activityRouter := router.PathPrefix("/activities").Subrouter()
	activityRouter.Use(middleware.AuthMiddleware)

	activityRouter.HandleFunc("", app.ActivityHandler.ListActivities).Methods("GET")
	activityRouter.HandleFunc("", app.ActivityHandler.CreateActivity).Methods("POST")
	activityRouter.HandleFunc("/batch", app.ActivityHandler.BatchCreateActivities).Methods("POST")
	activityRouter.HandleFunc("/batch", app.ActivityHandler.BatchDeleteActivities).Methods("DELETE")
	activityRouter.HandleFunc("/stats", app.ActivityHandler.GetStats).Methods("GET")
	activityRouter.HandleFunc("/{id}", app.ActivityHandler.GetActivity).Methods("GET")
	activityRouter.HandleFunc("/{id}", app.ActivityHandler.UpdateActivity).Methods("PATCH")
	activityRouter.HandleFunc("/{id}", app.ActivityHandler.DeleteActivity).Methods("DELETE")
	activityRouter.HandleFunc("/{id}/photos", app.photoHandler.Upload).Methods("POST")
	activityRouter.HandleFunc("/{id}/photos", app.photoHandler.GetActivityPhoto).Methods("GET")
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

// registerFeaturesRoutes registers the feature flags endpoint
func (app *Application) registerFeaturesRoutes(router *mux.Router) {
	featuresRouter := router.PathPrefix("/features").Subrouter()
	featuresRouter.Use(middleware.AuthMiddleware)
	featuresRouter.HandleFunc("", app.FeaturesHandler.GetFeatures).Methods("GET")
}

// registerWebhookRoutes registers webhook management routes
func (app *Application) registerWebhookRoutes(router *mux.Router) {
	webhookRouter := router.PathPrefix("/webhooks").Subrouter()
	webhookRouter.Use(middleware.AuthMiddleware)
	webhookRouter.HandleFunc("", app.WebhookHandler.CreateWebhook).Methods("POST")
	webhookRouter.HandleFunc("", app.WebhookHandler.ListWebhooks).Methods("GET")
	webhookRouter.HandleFunc("/{id}", app.WebhookHandler.DeleteWebhook).Methods("DELETE")
}

// registerExportRoutes registers export and job routes
func (app *Application) registerExportRoutes(router *mux.Router) {
	exportRouter := router.PathPrefix("/activities/export").Subrouter()
	exportRouter.Use(middleware.AuthMiddleware)
	exportRouter.HandleFunc("/csv", app.ExportHandler.ExportCSV).Methods("GET")
	exportRouter.HandleFunc("/pdf", app.ExportHandler.EnqueuePDFExport).Methods("POST")

	jobRouter := router.PathPrefix("/jobs").Subrouter()
	jobRouter.Use(middleware.AuthMiddleware)
	jobRouter.HandleFunc("/{jobId}/status", app.ExportHandler.GetJobStatus).Methods("GET")
	jobRouter.HandleFunc("/{jobId}/download", app.ExportHandler.GetDownloadURL).Methods("GET")
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

	// Start WebSocket hub event loop
	go app.WSHub.Run()

	// Subscribe webhook delivery to webhook bus
	webhookCtx, webhookCancel := context.WithCancel(context.Background())
	defer webhookCancel()
	if err := app.WebhookBus.Subscribe(webhookCtx, app.WebhookDelivery.Handle); err != nil {
		log.Printf("Warning: Failed to subscribe webhook delivery: %v", err)
	}

	app.WebhookRetryWorker.Start(webhookCtx)

	// Start scheduler
	app.Scheduler.Start()

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

	// Stop scheduler
	log.Println("⏳ Stopping scheduler...")
	app.Scheduler.Stop()
	log.Println("✅ Scheduler stopped")

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
