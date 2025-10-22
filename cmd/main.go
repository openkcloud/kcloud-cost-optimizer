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
	"github.com/kcloud-opt/policy/internal/types"
	"github.com/kcloud-opt/policy/internal/validator"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
	goVersion = "unknown"
)

// LoggerWrapper wraps logger.Logger to implement types.Logger interface
type LoggerWrapper struct {
	*logger.Logger
}

func (l *LoggerWrapper) Info(msg string, fields ...interface{}) {
	l.Logger.Info(msg)
}

func (l *LoggerWrapper) Warn(msg string, fields ...interface{}) {
	l.Logger.Warn(msg)
}

func (l *LoggerWrapper) Error(msg string, fields ...interface{}) {
	l.Logger.Error(msg)
}

func (l *LoggerWrapper) Debug(msg string, fields ...interface{}) {
	l.Logger.Debug(msg)
}

func (l *LoggerWrapper) Fatal(msg string, fields ...interface{}) {
	l.Logger.Fatal(msg)
}

func (l *LoggerWrapper) WithError(err error) types.Logger {
	return l
}

func (l *LoggerWrapper) WithDuration(duration time.Duration) types.Logger {
	return l
}

func (l *LoggerWrapper) WithPolicy(policyID, policyName string) types.Logger {
	return l
}

func (l *LoggerWrapper) WithWorkload(workloadID, workloadType string) types.Logger {
	return l
}

func (l *LoggerWrapper) WithEvaluation(evaluationID string) types.Logger {
	return l
}

func main() {
	// Initialize logger
	loggerInstance, err := logger.NewLogger(&config.LoggingConfig{
		Level:  "info",
		Format: "json",
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	loggerInstance.Info("Starting Policy Engine")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		loggerInstance.Fatal("Failed to load configuration")
	}

	loggerInstance.Info("Configuration loaded")

	// Create types.Logger interface wrapper
	var appLogger types.Logger = &LoggerWrapper{loggerInstance}

	// Initialize metrics
	metricsInstance := metrics.NewMetrics(appLogger)
	metricsInstance.Initialize()

	// Initialize storage
	storageManager := memory.NewStorageManager()
	loggerInstance.Info("Storage manager initialized")

	// Initialize validator
	validationEngine := validator.NewValidationEngine(appLogger)
	if err := validationEngine.Initialize(context.Background()); err != nil {
		loggerInstance.Fatal("Failed to initialize validation engine")
	}
	loggerInstance.Info("Validation engine initialized")

	// Initialize evaluator components
	ruleEngine := evaluator.NewRuleEngine(appLogger)
	policyEvaluator := evaluator.NewPolicyEvaluator(storageManager, ruleEngine, appLogger)
	conflictResolver := evaluator.NewConflictResolver(appLogger)

	evaluationEngine := evaluator.NewEvaluationEngine(policyEvaluator, conflictResolver, storageManager, appLogger)
	loggerInstance.Info("Evaluation engine initialized")

	// Initialize automation engine
	automationEngine := automation.NewAutomationEngine(storageManager, nil, nil, nil, appLogger)
	if err := automationEngine.Initialize(context.Background()); err != nil {
		loggerInstance.Fatal("Failed to initialize automation engine")
	}
	loggerInstance.Info("Automation engine initialized")

	// Initialize handlers
	handlersInstance := handlers.NewHandlers(storageManager, evaluationEngine, automationEngine, appLogger)
	loggerInstance.Info("Handlers initialized")

	// Initialize router
	router := routes.NewRouter(handlersInstance, cfg, loggerInstance)
	httpRouter := router.SetupRoutes()
	loggerInstance.Info("Router initialized")

	// Start metrics collection
	metricsManager := metrics.NewMetricsManager(metricsInstance, appLogger)
	go metricsManager.Start(context.Background())
	loggerInstance.Info("Metrics collection started")

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
		loggerInstance.Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loggerInstance.Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	loggerInstance.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		loggerInstance.Error("Server forced to shutdown")
	}

	loggerInstance.Info("Server exited")
}
