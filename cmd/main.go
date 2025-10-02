package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kcloud-opt/policy/api/handlers"
	"github.com/kcloud-opt/policy/api/routes"
	"github.com/kcloud-opt/policy/internal/automation"
	"github.com/kcloud-opt/policy/internal/config"
	"github.com/kcloud-opt/policy/internal/evaluator"
	"github.com/kcloud-opt/policy/internal/logger"
	"github.com/kcloud-opt/policy/internal/metrics"
	"github.com/kcloud-opt/policy/internal/storage/memory"
	"github.com/kcloud-opt/policy/internal/validator"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
	goVersion = "unknown"
)

func main() {
	// Initialize logger
	logger, err := logger.NewLogger(&config.LogConfig{
		Level:    "info",
		Encoding: "json",
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.Info("Starting Policy Engine",
		"version", version,
		"build_time", buildTime,
		"git_commit", gitCommit,
		"go_version", goVersion,
	)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
	}

	logger.Info("Configuration loaded", "config", cfg)

	// Initialize metrics
	metricsInstance := metrics.NewMetrics(logger)
	metricsInstance.Initialize()

	// Initialize storage
	storageManager := memory.NewStorageManager()
	logger.Info("Storage manager initialized")

	// Initialize validator
	validationEngine := validator.NewValidationEngine(logger)
	if err := validationEngine.Initialize(context.Background()); err != nil {
		logger.Fatal("Failed to initialize validation engine", "error", err)
	}
	logger.Info("Validation engine initialized")

	// Initialize evaluator
	evaluationEngine := evaluator.NewEvaluationEngine(storageManager, logger)
	logger.Info("Evaluation engine initialized")

	// Initialize automation engine
	automationEngine := automation.NewAutomationEngine(storageManager, logger)
	if err := automationEngine.Initialize(context.Background()); err != nil {
		logger.Fatal("Failed to initialize automation engine", "error", err)
	}
	logger.Info("Automation engine initialized")

	// Initialize handlers
	handlersInstance := handlers.NewHandlers(storageManager, evaluationEngine, automationEngine, logger)
	logger.Info("Handlers initialized")

	// Initialize router
	router := routes.NewRouter(handlersInstance, cfg, logger)
	httpRouter := router.SetupRoutes()
	logger.Info("Router initialized")

	// Start metrics collection
	metricsManager := metrics.NewMetricsManager(metricsInstance, logger)
	go metricsManager.Start(context.Background())
	logger.Info("Metrics collection started")

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      httpRouter,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	// Shutdown automation engine
	if err := automationEngine.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown automation engine", "error", err)
	}

	logger.Info("Server exited")
}
