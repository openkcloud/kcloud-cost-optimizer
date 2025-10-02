package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kcloud-opt/policy/api/handlers"
	"github.com/kcloud-opt/policy/api/routes"
	"github.com/kcloud-opt/policy/internal/automation"
	"github.com/kcloud-opt/policy/internal/config"
	"github.com/kcloud-opt/policy/internal/evaluator"
	"github.com/kcloud-opt/policy/internal/logger"
	"github.com/kcloud-opt/policy/internal/storage/memory"
	"github.com/kcloud-opt/policy/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer represents a test server instance
type TestServer struct {
	Server     *httptest.Server
	Handlers   *handlers.Handlers
	Storage    *memory.StorageManager
	Evaluator  *evaluator.EvaluationEngine
	Automation *automation.AutomationEngine
}

// SetupTestServer creates a test server with all components initialized
func SetupTestServer(t *testing.T) *TestServer {
	// Initialize logger
	testLogger, err := logger.NewLogger(&config.LogConfig{
		Level:    "debug",
		Encoding: "console",
	})
	require.NoError(t, err)

	// Initialize storage
	storageManager := memory.NewStorageManager()

	// Initialize validator
	validationEngine := validator.NewValidationEngine(testLogger)
	err = validationEngine.Initialize(context.Background())
	require.NoError(t, err)

	// Initialize evaluator
	evaluationEngine := evaluator.NewEvaluationEngine(storageManager, testLogger)

	// Initialize automation engine
	automationEngine := automation.NewAutomationEngine(storageManager, nil, nil, nil, testLogger)
	err = automationEngine.Initialize(context.Background())
	require.NoError(t, err)

	// Initialize handlers
	handlersInstance := handlers.NewHandlers(storageManager, evaluationEngine, automationEngine, testLogger)

	// Initialize router
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
	}
	router := routes.NewRouter(handlersInstance, cfg, testLogger)
	httpRouter := router.SetupRoutes()

	// Create test server
	server := httptest.NewServer(httpRouter)

	return &TestServer{
		Server:     server,
		Handlers:   handlersInstance,
		Storage:    storageManager,
		Evaluator:  evaluationEngine,
		Automation: automationEngine,
	}
}

// CleanupTestServer cleans up the test server
func (ts *TestServer) CleanupTestServer() {
	if ts.Server != nil {
		ts.Server.Close()
	}
	if ts.Automation != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ts.Automation.Shutdown(ctx)
	}
}

// TestIntegrationBasicFlow tests the basic policy engine flow
func TestIntegrationBasicFlow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := client.Get(ts.Server.URL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)
		assert.Equal(t, "healthy", health["status"])
	})

	t.Run("CreatePolicy", func(t *testing.T) {
		policyData := map[string]interface{}{
			"name":        "test-policy",
			"description": "Test policy for integration testing",
			"type":        "cost-optimization",
			"enabled":     true,
			"rules": []map[string]interface{}{
				{
					"name":        "cpu-rule",
					"description": "CPU optimization rule",
					"condition":   "cpu_usage > 80",
					"action":      "scale_up",
					"priority":    1,
				},
			},
		}

		jsonData, err := json.Marshal(policyData)
		require.NoError(t, err)

		resp, err := client.Post(ts.Server.URL+"/api/v1/policies", "application/json", bytes.NewReader(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.NotEmpty(t, result["id"])
	})

	t.Run("CreateWorkload", func(t *testing.T) {
		workloadData := map[string]interface{}{
			"name":        "test-workload",
			"description": "Test workload for integration testing",
			"namespace":   "default",
			"resources": map[string]interface{}{
				"cpu":    "100m",
				"memory": "128Mi",
			},
			"labels": map[string]string{
				"app": "test-app",
			},
		}

		jsonData, err := json.Marshal(workloadData)
		require.NoError(t, err)

		resp, err := client.Post(ts.Server.URL+"/api/v1/workloads", "application/json", bytes.NewReader(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.NotEmpty(t, result["id"])
	})

	t.Run("EvaluateWorkload", func(t *testing.T) {
		// First create a workload
		workloadData := map[string]interface{}{
			"name":        "evaluation-workload",
			"description": "Workload for evaluation testing",
			"namespace":   "default",
			"resources": map[string]interface{}{
				"cpu":    "200m",
				"memory": "256Mi",
			},
		}

		jsonData, err := json.Marshal(workloadData)
		require.NoError(t, err)

		resp, err := client.Post(ts.Server.URL+"/api/v1/workloads", "application/json", bytes.NewReader(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var workloadResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&workloadResult)
		require.NoError(t, err)
		workloadID := workloadResult["id"].(string)

		// Now evaluate the workload
		resp, err = client.Post(ts.Server.URL+"/api/v1/evaluations/workload/"+workloadID, "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var evalResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&evalResult)
		require.NoError(t, err)
		assert.NotEmpty(t, evalResult["evaluation_id"])
	})

	t.Run("CreateAutomationRule", func(t *testing.T) {
		ruleData := map[string]interface{}{
			"name":        "test-automation-rule",
			"description": "Test automation rule for integration testing",
			"enabled":     true,
			"trigger": map[string]interface{}{
				"type": "schedule",
				"schedule": map[string]interface{}{
					"interval": "5m",
				},
			},
			"condition": "true",
			"action": map[string]interface{}{
				"type": "log",
				"params": map[string]interface{}{
					"message": "Automated action executed",
				},
			},
		}

		jsonData, err := json.Marshal(ruleData)
		require.NoError(t, err)

		resp, err := client.Post(ts.Server.URL+"/api/v1/automation/rules", "application/json", bytes.NewReader(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.NotEmpty(t, result["id"])
	})

	t.Run("GetMetrics", func(t *testing.T) {
		resp, err := client.Get(ts.Server.URL + "/metrics")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "policy_engine_")
	})
}

// TestIntegrationWithYAMLFiles tests integration with actual YAML files
func TestIntegrationWithYAMLFiles(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Test with example policy files
	policyFiles := []string{
		"cost-optimization-policy.yaml",
		"automation-rule.yaml",
		"workload-priority-policy.yaml",
	}

	for _, policyFile := range policyFiles {
		t.Run(fmt.Sprintf("LoadPolicy_%s", policyFile), func(t *testing.T) {
			filePath := filepath.Join("..", "examples", "policies", policyFile)

			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Skipf("Policy file %s does not exist", filePath)
				return
			}

			// Read file content
			content, err := os.ReadFile(filePath)
			require.NoError(t, err)

			// Create policy
			resp, err := client.Post(ts.Server.URL+"/api/v1/policies", "application/yaml", bytes.NewReader(content))
			require.NoError(t, err)
			defer resp.Body.Close()

			// Should create successfully or return validation error
			assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusBadRequest)
		})
	}

	// Test with example workload file
	t.Run("LoadWorkload", func(t *testing.T) {
		filePath := filepath.Join("..", "examples", "workloads", "sample-workload.yaml")

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Skipf("Workload file %s does not exist", filePath)
			return
		}

		// Read file content
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)

		// Create workload
		resp, err := client.Post(ts.Server.URL+"/api/v1/workloads", "application/yaml", bytes.NewReader(content))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should create successfully or return validation error
		assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusBadRequest)
	})
}

// TestIntegrationErrorHandling tests error handling scenarios
func TestIntegrationErrorHandling(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	t.Run("InvalidPolicyData", func(t *testing.T) {
		invalidData := []byte(`{"invalid": "json"`)

		resp, err := client.Post(ts.Server.URL+"/api/v1/policies", "application/json", bytes.NewReader(invalidData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("NonExistentPolicy", func(t *testing.T) {
		resp, err := client.Get(ts.Server.URL + "/api/v1/policies/non-existent")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("NonExistentWorkload", func(t *testing.T) {
		resp, err := client.Get(ts.Server.URL + "/api/v1/workloads/non-existent")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("InvalidEvaluationRequest", func(t *testing.T) {
		resp, err := client.Post(ts.Server.URL+"/api/v1/evaluations/workload/non-existent", "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestIntegrationPerformance tests basic performance characteristics
func TestIntegrationPerformance(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	t.Run("ConcurrentRequests", func(t *testing.T) {
		const numRequests = 10
		done := make(chan bool, numRequests)

		// Send concurrent health check requests
		for i := 0; i < numRequests; i++ {
			go func() {
				defer func() { done <- true }()

				resp, err := client.Get(ts.Server.URL + "/health")
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}()
		}

		// Wait for all requests to complete
		for i := 0; i < numRequests; i++ {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}
	})

	t.Run("ResponseTime", func(t *testing.T) {
		start := time.Now()

		resp, err := client.Get(ts.Server.URL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		duration := time.Since(start)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Health check should respond within 100ms
		assert.Less(t, duration, 100*time.Millisecond)
	})
}

// TestIntegrationDataPersistence tests data persistence across requests
func TestIntegrationDataPersistence(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	t.Run("PolicyPersistence", func(t *testing.T) {
		// Create policy
		policyData := map[string]interface{}{
			"name":        "persistence-test-policy",
			"description": "Policy for testing data persistence",
			"type":        "cost-optimization",
			"enabled":     true,
		}

		jsonData, err := json.Marshal(policyData)
		require.NoError(t, err)

		resp, err := client.Post(ts.Server.URL+"/api/v1/policies", "application/json", bytes.NewReader(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResult)
		require.NoError(t, err)
		policyID := createResult["id"].(string)

		// Retrieve policy
		resp, err = client.Get(ts.Server.URL + "/api/v1/policies/" + policyID)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var getResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&getResult)
		require.NoError(t, err)
		assert.Equal(t, policyID, getResult["id"])
		assert.Equal(t, "persistence-test-policy", getResult["name"])
	})
}

// TestIntegrationWorkflow tests a complete workflow
func TestIntegrationWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer()

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Step 1: Create a policy
		policyData := map[string]interface{}{
			"name":        "workflow-test-policy",
			"description": "Policy for complete workflow testing",
			"type":        "cost-optimization",
			"enabled":     true,
			"rules": []map[string]interface{}{
				{
					"name":        "memory-rule",
					"description": "Memory optimization rule",
					"condition":   "memory_usage > 90",
					"action":      "scale_down",
					"priority":    1,
				},
			},
		}

		jsonData, err := json.Marshal(policyData)
		require.NoError(t, err)

		resp, err := client.Post(ts.Server.URL+"/api/v1/policies", "application/json", bytes.NewReader(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var policyResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&policyResult)
		require.NoError(t, err)
		policyID := policyResult["id"].(string)

		// Step 2: Create a workload
		workloadData := map[string]interface{}{
			"name":        "workflow-test-workload",
			"description": "Workload for complete workflow testing",
			"namespace":   "default",
			"resources": map[string]interface{}{
				"cpu":    "500m",
				"memory": "1Gi",
			},
		}

		jsonData, err = json.Marshal(workloadData)
		require.NoError(t, err)

		resp, err = client.Post(ts.Server.URL+"/api/v1/workloads", "application/json", bytes.NewReader(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var workloadResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&workloadResult)
		require.NoError(t, err)
		workloadID := workloadResult["id"].(string)

		// Step 3: Evaluate the workload against the policy
		resp, err = client.Post(ts.Server.URL+"/api/v1/evaluations/workload/"+workloadID, "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var evalResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&evalResult)
		require.NoError(t, err)
		assert.NotEmpty(t, evalResult["evaluation_id"])

		// Step 4: Create an automation rule
		ruleData := map[string]interface{}{
			"name":        "workflow-automation-rule",
			"description": "Automation rule for workflow testing",
			"enabled":     true,
			"trigger": map[string]interface{}{
				"type": "event",
				"event": map[string]interface{}{
					"name": "policy_violation",
				},
			},
			"condition": "workload_id == \"" + workloadID + "\"",
			"action": map[string]interface{}{
				"type": "notification",
				"params": map[string]interface{}{
					"message": "Policy violation detected for workload " + workloadID,
				},
			},
		}

		jsonData, err = json.Marshal(ruleData)
		require.NoError(t, err)

		resp, err = client.Post(ts.Server.URL+"/api/v1/automation/rules", "application/json", bytes.NewReader(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var ruleResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&ruleResult)
		require.NoError(t, err)
		ruleID := ruleResult["id"].(string)

		// Step 5: Verify all resources exist
		resp, err = client.Get(ts.Server.URL + "/api/v1/policies/" + policyID)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = client.Get(ts.Server.URL + "/api/v1/workloads/" + workloadID)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = client.Get(ts.Server.URL + "/api/v1/automation/rules/" + ruleID)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Step 6: Clean up - delete resources
		resp, err = client.Get(ts.Server.URL + "/api/v1/policies/" + policyID)
		require.NoError(t, err)
		defer resp.Body.Close()
		req, err := http.NewRequest(http.MethodDelete, ts.Server.URL+"/api/v1/policies/"+policyID, nil)
		require.NoError(t, err)

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		req, err = http.NewRequest(http.MethodDelete, ts.Server.URL+"/api/v1/workloads/"+workloadID, nil)
		require.NoError(t, err)

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		req, err = http.NewRequest(http.MethodDelete, ts.Server.URL+"/api/v1/automation/rules/"+ruleID, nil)
		require.NoError(t, err)

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
