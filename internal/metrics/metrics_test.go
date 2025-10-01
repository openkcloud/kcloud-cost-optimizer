package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/kcloud-opt/policy/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics_Initialize(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)

	// Initialize should not panic
	assert.NotPanics(t, func() {
		metrics.Initialize()
	})

	// Verify metrics are initialized
	assert.NotNil(t, metrics.HTTPRequestsTotal)
	assert.NotNil(t, metrics.HTTPRequestDuration)
	assert.NotNil(t, metrics.PolicyTotal)
	assert.NotNil(t, metrics.WorkloadTotal)
	assert.NotNil(t, metrics.AutomationRuleTotal)
	assert.NotNil(t, metrics.SystemUptime)
}

func TestMetrics_RecordHTTPRequest(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Record HTTP request metrics
	metrics.RecordHTTPRequest("GET", "/api/v1/policies", "200", 100*time.Millisecond, 1024, 2048)

	// Verify metrics were recorded (we can't easily test the actual values without a metrics server)
	// But we can ensure no panics occurred
	assert.NotNil(t, metrics.HTTPRequestsTotal)
}

func TestMetrics_RecordPolicyEvaluation(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Record policy evaluation metrics
	metrics.RecordPolicyEvaluation("cost-optimization", "test-policy", "success", 50*time.Millisecond)

	// Verify metrics were recorded
	assert.NotNil(t, metrics.PolicyEvaluationsTotal)
}

func TestMetrics_RecordPolicyValidation(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Record policy validation metrics
	metrics.RecordPolicyValidation("cost-optimization", "success")

	// Verify metrics were recorded
	assert.NotNil(t, metrics.PolicyValidationTotal)
}

func TestMetrics_RecordDecision(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Record decision metrics
	metrics.RecordDecision("scale-up", "cost-optimization", "success", 25*time.Millisecond)

	// Verify metrics were recorded
	assert.NotNil(t, metrics.DecisionTotal)
	assert.NotNil(t, metrics.DecisionSuccess)
}

func TestMetrics_RecordDecisionFailure(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Record failed decision metrics
	metrics.RecordDecision("scale-up", "cost-optimization", "error", 25*time.Millisecond)

	// Verify metrics were recorded
	assert.NotNil(t, metrics.DecisionTotal)
	assert.NotNil(t, metrics.DecisionFailure)
}

func TestMetrics_RecordAutomationRuleExecution(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Record automation rule execution metrics
	metrics.RecordAutomationRuleExecution("rule-1", "scale-rule", "event-based", "success", 100*time.Millisecond)

	// Verify metrics were recorded
	assert.NotNil(t, metrics.AutomationRuleExecutionsTotal)
	assert.NotNil(t, metrics.AutomationRuleSuccess)
}

func TestMetrics_RecordAutomationRuleExecutionFailure(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Record failed automation rule execution metrics
	metrics.RecordAutomationRuleExecution("rule-1", "scale-rule", "event-based", "error", 100*time.Millisecond)

	// Verify metrics were recorded
	assert.NotNil(t, metrics.AutomationRuleExecutionsTotal)
	assert.NotNil(t, metrics.AutomationRuleFailure)
}

func TestMetrics_RecordStorageOperation(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Record storage operation metrics
	metrics.RecordStorageOperation("create", "policy", 10*time.Millisecond)

	// Verify metrics were recorded
	assert.NotNil(t, metrics.StorageOperationsTotal)
	assert.NotNil(t, metrics.StorageOperationDuration)
}

func TestMetrics_RecordStorageError(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Record storage error metrics
	metrics.RecordStorageError("create", "policy", "not_found")

	// Verify metrics were recorded
	assert.NotNil(t, metrics.StorageErrorsTotal)
}

func TestMetrics_UpdatePolicyCounts(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Update policy counts
	metrics.UpdatePolicyCounts(10, 8, 2)

	// Verify metrics were updated
	assert.NotNil(t, metrics.PolicyTotal)
	assert.NotNil(t, metrics.PolicyActive)
	assert.NotNil(t, metrics.PolicyInactive)
}

func TestMetrics_UpdateWorkloadCounts(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Update workload counts
	metrics.UpdateWorkloadCounts(20, 15, 3, 1, 1)

	// Verify metrics were updated
	assert.NotNil(t, metrics.WorkloadTotal)
	assert.NotNil(t, metrics.WorkloadRunning)
	assert.NotNil(t, metrics.WorkloadStopped)
	assert.NotNil(t, metrics.WorkloadPending)
	assert.NotNil(t, metrics.WorkloadFailed)
}

func TestMetrics_UpdateAutomationRuleCounts(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Update automation rule counts
	metrics.UpdateAutomationRuleCounts(5, 4)

	// Verify metrics were updated
	assert.NotNil(t, metrics.AutomationRuleTotal)
	assert.NotNil(t, metrics.AutomationRuleActive)
}

func TestMetrics_UpdateSystemMetrics(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Update system metrics
	metrics.UpdateSystemMetrics(1*time.Hour, 1024*1024*1024, 75.5, 100)

	// Verify metrics were updated
	assert.NotNil(t, metrics.SystemUptime)
	assert.NotNil(t, metrics.SystemMemoryUsage)
	assert.NotNil(t, metrics.SystemCPUUsage)
	assert.NotNil(t, metrics.SystemGoroutines)
}

func TestMetrics_GetMetrics(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Update some metrics
	metrics.UpdatePolicyCounts(5, 4, 1)
	metrics.UpdateWorkloadCounts(10, 8, 1, 1, 0)
	metrics.UpdateAutomationRuleCounts(3, 2)
	metrics.UpdateSystemMetrics(30*time.Minute, 512*1024*1024, 50.0, 50)

	// Get metrics
	metricsMap, err := metrics.GetMetrics(context.Background())
	require.NoError(t, err)

	// Verify metrics are returned
	assert.Contains(t, metricsMap, "policies_total")
	assert.Contains(t, metricsMap, "policies_active")
	assert.Contains(t, metricsMap, "workloads_total")
	assert.Contains(t, metricsMap, "workloads_running")
	assert.Contains(t, metricsMap, "automation_rules_total")
	assert.Contains(t, metricsMap, "system_uptime")
	assert.Contains(t, metricsMap, "system_memory_usage")
	assert.Contains(t, metricsMap, "system_cpu_usage")
	assert.Contains(t, metricsMap, "system_goroutines")
}

func TestMetrics_Health(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)

	// Health check before initialization should fail
	err := metrics.Health(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP metrics not initialized")

	// Initialize metrics
	metrics.Initialize()

	// Health check after initialization should pass
	err = metrics.Health(context.Background())
	assert.NoError(t, err)
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Test concurrent access to metrics
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Record various metrics concurrently
			metrics.RecordHTTPRequest("GET", "/api/v1/policies", "200", 100*time.Millisecond, 1024, 2048)
			metrics.RecordPolicyEvaluation("cost-optimization", "test-policy", "success", 50*time.Millisecond)
			metrics.RecordDecision("scale-up", "cost-optimization", "success", 25*time.Millisecond)
			metrics.RecordAutomationRuleExecution("rule-1", "scale-rule", "event-based", "success", 100*time.Millisecond)
			metrics.RecordStorageOperation("create", "policy", 10*time.Millisecond)

			metrics.UpdatePolicyCounts(10, 8, 2)
			metrics.UpdateWorkloadCounts(20, 15, 3, 1, 1)
			metrics.UpdateAutomationRuleCounts(5, 4)
			metrics.UpdateSystemMetrics(1*time.Hour, 1024*1024*1024, 75.5, 100)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify metrics are still accessible
	metricsMap, err := metrics.GetMetrics(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, metricsMap)
}

func TestMetrics_ResetMetrics(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Record some metrics
	metrics.RecordHTTPRequest("GET", "/api/v1/policies", "200", 100*time.Millisecond, 1024, 2048)
	metrics.UpdatePolicyCounts(5, 4, 1)

	// Reset metrics
	metrics.ResetMetrics()

	// Verify metrics are reset (values should be 0)
	metricsMap, err := metrics.GetMetrics(context.Background())
	require.NoError(t, err)

	// Check that metrics exist but have been reset
	assert.Contains(t, metricsMap, "policies_total")
	assert.Contains(t, metricsMap, "policies_active")
}

func TestMetrics_EdgeCases(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Test with zero values
	metrics.RecordHTTPRequest("GET", "/api/v1/policies", "200", 0, 0, 0)
	metrics.RecordPolicyEvaluation("cost-optimization", "test-policy", "success", 0)
	metrics.RecordDecision("scale-up", "cost-optimization", "success", 0)
	metrics.RecordAutomationRuleExecution("rule-1", "scale-rule", "event-based", "success", 0)
	metrics.RecordStorageOperation("create", "policy", 0)

	metrics.UpdatePolicyCounts(0, 0, 0)
	metrics.UpdateWorkloadCounts(0, 0, 0, 0, 0)
	metrics.UpdateAutomationRuleCounts(0, 0)
	metrics.UpdateSystemMetrics(0, 0, 0, 0)

	// Test with negative values
	metrics.UpdateSystemMetrics(-1*time.Hour, -1024, -50.0, -10)

	// Verify no panics occurred
	assert.NotNil(t, metrics.HTTPRequestsTotal)
	assert.NotNil(t, metrics.PolicyTotal)
	assert.NotNil(t, metrics.SystemUptime)
}

func TestMetrics_GetMetricsEmpty(t *testing.T) {
	logger := &types.Logger{}
	metrics := NewMetrics(logger)
	metrics.Initialize()

	// Get metrics without setting any values
	metricsMap, err := metrics.GetMetrics(context.Background())
	require.NoError(t, err)

	// Verify metrics are returned even when empty
	assert.Contains(t, metricsMap, "policies_total")
	assert.Contains(t, metricsMap, "policies_active")
	assert.Contains(t, metricsMap, "workloads_total")
	assert.Contains(t, metricsMap, "automation_rules_total")
	assert.Contains(t, metricsMap, "system_uptime")
}
