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

	"go.uber.org/zap"

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

// convertFields converts interface{} fields to zap.Field format
// Supports key-value pairs: ("key1", value1, "key2", value2, ...)
// Or zap.Field directly
func convertFields(fields ...interface{}) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zap.Field, 0, len(fields))
	
	// Process fields as key-value pairs
	for i := 0; i < len(fields); i++ {
		// Check if it's already a zap.Field
		if field, ok := fields[i].(zap.Field); ok {
			zapFields = append(zapFields, field)
			continue
		}
		
		// Try to process as key-value pair
		if i < len(fields)-1 {
			key, ok := fields[i].(string)
			if ok {
				value := fields[i+1]
				zapFields = append(zapFields, zap.Any(key, value))
				i++ // Skip next field as we've already processed it
				continue
			}
		}
		
		// If not a key-value pair or zap.Field, add as Any
		zapFields = append(zapFields, zap.Any(fmt.Sprintf("field_%d", i), fields[i]))
	}
	
	return zapFields
}

func (l *LoggerWrapper) Info(msg string, fields ...interface{}) {
	zapFields := convertFields(fields...)
	l.Logger.Info(msg, zapFields...)
}

func (l *LoggerWrapper) Warn(msg string, fields ...interface{}) {
	zapFields := convertFields(fields...)
	l.Logger.Warn(msg, zapFields...)
}

func (l *LoggerWrapper) Error(msg string, fields ...interface{}) {
	zapFields := convertFields(fields...)
	l.Logger.Error(msg, zapFields...)
}

func (l *LoggerWrapper) Debug(msg string, fields ...interface{}) {
	zapFields := convertFields(fields...)
	l.Logger.Debug(msg, zapFields...)
}

func (l *LoggerWrapper) Fatal(msg string, fields ...interface{}) {
	zapFields := convertFields(fields...)
	l.Logger.Fatal(msg, zapFields...)
}

func (l *LoggerWrapper) WithError(err error) types.Logger {
	return &LoggerWrapper{l.Logger.WithError(err)}
}

func (l *LoggerWrapper) WithDuration(duration time.Duration) types.Logger {
	return &LoggerWrapper{l.Logger.WithDuration(duration)}
}

func (l *LoggerWrapper) WithPolicy(policyID, policyName string) types.Logger {
	return &LoggerWrapper{l.Logger.WithPolicy(policyID, policyName)}
}

func (l *LoggerWrapper) WithWorkload(workloadID, workloadType string) types.Logger {
	return &LoggerWrapper{l.Logger.WithWorkload(workloadID, workloadType)}
}

func (l *LoggerWrapper) WithEvaluation(evaluationID string) types.Logger {
	return &LoggerWrapper{l.Logger.WithEvaluation(evaluationID)}
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

	// Initialize storage (required - core functionality)
	storageManager := memory.NewStorageManager()
	loggerInstance.Info("Storage manager initialized")

	// Initialize metrics (optional - monitoring feature, but initialization always succeeds)
	metricsInstance := metrics.NewMetrics(appLogger)
	metricsInstance.Initialize()
	loggerInstance.Info("Metrics initialized")

	// Initialize validator (partially optional - some features may be limited)
	validationEngine := validator.NewValidationEngine(appLogger)
	if err := validationEngine.Initialize(context.Background()); err != nil {
		loggerInstance.WithError(err).Warn("Failed to initialize validation engine - continuing with limited validation")
		validationEngine = nil // Disable validation if initialization fails
	} else {
		loggerInstance.Info("Validation engine initialized")
	}

	// Initialize evaluator components (required - core functionality)
	ruleEngine := evaluator.NewRuleEngine(appLogger)
	policyEvaluator := evaluator.NewPolicyEvaluator(storageManager, ruleEngine, appLogger)
	conflictResolver := evaluator.NewConflictResolver(appLogger)

	evaluationEngine := evaluator.NewEvaluationEngine(policyEvaluator, conflictResolver, storageManager, appLogger)
	loggerInstance.Info("Evaluation engine initialized")

	// Initialize automation engine (optional - automation feature)
	var automationEngine automation.AutomationEngine
	if ae := automation.NewAutomationEngine(storageManager, nil, nil, nil, appLogger); ae != nil {
		if err := ae.Initialize(context.Background()); err != nil {
			loggerInstance.WithError(err).Warn("Failed to initialize automation engine - continuing without automation")
			automationEngine = nil // Disable automation if initialization fails
		} else {
			automationEngine = ae
			loggerInstance.Info("Automation engine initialized")
		}
	} else {
		loggerInstance.Warn("Automation engine not available - continuing without automation")
		automationEngine = nil
	}

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
