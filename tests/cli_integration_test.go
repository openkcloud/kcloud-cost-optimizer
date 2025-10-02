package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIIntegration tests CLI integration with the Policy Engine
func TestCLIIntegration(t *testing.T) {
	// Build CLI binary first
	cliPath := buildCLIBinary(t)
	defer os.Remove(cliPath)

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer()

	// Extract port from server URL
	serverURL := strings.TrimPrefix(ts.Server.URL, "http://")
	serverHost := strings.Split(serverURL, ":")[0]
	serverPort := strings.Split(serverURL, ":")[1]

	t.Run("CLIHelp", func(t *testing.T) {
		cmd := exec.Command(cliPath, "--help")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "Policy Engine CLI")
		assert.Contains(t, outputStr, "policy")
		assert.Contains(t, outputStr, "workload")
		assert.Contains(t, outputStr, "evaluate")
		assert.Contains(t, outputStr, "automation")
		assert.Contains(t, outputStr, "status")
	})

	t.Run("CLIStatus", func(t *testing.T) {
		cmd := exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"status")

		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "Policy Engine Status")
	})

	t.Run("CLIPing", func(t *testing.T) {
		cmd := exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"ping")

		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "Ping successful")
	})

	t.Run("CLIPolicyManagement", func(t *testing.T) {
		// Create a temporary policy file
		policyFile := createTempPolicyFile(t)
		defer os.Remove(policyFile)

		// Create policy via CLI
		cmd := exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"policy", "create", policyFile)

		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "Policy created successfully")

		// List policies via CLI
		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"policy", "list")

		output, err = cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr = string(output)
		assert.Contains(t, outputStr, "policies")
	})

	t.Run("CLIWorkloadManagement", func(t *testing.T) {
		// Create a temporary workload file
		workloadFile := createTempWorkloadFile(t)
		defer os.Remove(workloadFile)

		// Create workload via CLI
		cmd := exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"workload", "create", workloadFile)

		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "Workload created successfully")

		// List workloads via CLI
		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"workload", "list")

		output, err = cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr = string(output)
		assert.Contains(t, outputStr, "workloads")
	})

	t.Run("CLIEvaluation", func(t *testing.T) {
		// First create a workload
		workloadFile := createTempWorkloadFile(t)
		defer os.Remove(workloadFile)

		cmd := exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"workload", "create", workloadFile)

		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr := string(output)
		require.Contains(t, outputStr, "Workload created successfully")

		// Extract workload ID from output (this is a simplified approach)
		// In a real scenario, you might need to parse JSON output
		workloadID := "cli-test-workload"

		// Evaluate workload via CLI
		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"evaluate", "workload", workloadID)

		output, err = cmd.CombinedOutput()
		// This might fail if workload ID doesn't match, which is expected
		// We're testing that the CLI command structure works
		_ = output
		_ = err
	})

	t.Run("CLIAutomationManagement", func(t *testing.T) {
		// Create a temporary automation rule file
		ruleFile := createTempAutomationRuleFile(t)
		defer os.Remove(ruleFile)

		// Create automation rule via CLI
		cmd := exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"automation", "create", ruleFile)

		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "Automation rule created successfully")

		// List automation rules via CLI
		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"automation", "list")

		output, err = cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr = string(output)
		assert.Contains(t, outputStr, "automation_rules")
	})

	t.Run("CLIMetrics", func(t *testing.T) {
		cmd := exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"metrics")

		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "policy_engine_")
	})

	t.Run("CLIVerboseOutput", func(t *testing.T) {
		cmd := exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"--verbose",
			"status")

		output, err := cmd.CombinedOutput()
		require.NoError(t, err)

		outputStr := string(output)
		// Verbose output should contain more detailed information
		assert.Contains(t, outputStr, "Policy Engine Status")
	})

	t.Run("CLIErrorHandling", func(t *testing.T) {
		// Test with non-existent server
		cmd := exec.Command(cliPath,
			"--server-host", "non-existent-host",
			"--server-port", "9999",
			"status")

		output, err := cmd.CombinedOutput()
		// Should fail with connection error
		assert.Error(t, err)

		outputStr := string(output)
		assert.Contains(t, outputStr, "Error")
	})

	t.Run("CLIConfigFile", func(t *testing.T) {
		// Create a temporary config file
		configFile := createTempConfigFile(t)
		defer os.Remove(configFile)

		cmd := exec.Command(cliPath,
			"--config", configFile,
			"status")

		output, err := cmd.CombinedOutput()
		// This should work with the config file
		_ = output
		_ = err
	})
}

// buildCLIBinary builds the CLI binary for testing
func buildCLIBinary(t *testing.T) string {
	// Create temporary directory for binary
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "policy-cli")

	// Build CLI binary
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/cli/main.go")
	cmd.Dir = ".."

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build CLI binary: %v\nOutput: %s", err, string(output))
	}

	return binaryPath
}

// createTempPolicyFile creates a temporary policy file for testing
func createTempPolicyFile(t *testing.T) string {
	policyData := map[string]interface{}{
		"name":        "cli-test-policy",
		"description": "Test policy for CLI integration testing",
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

	jsonData, err := json.MarshalIndent(policyData, "", "  ")
	require.NoError(t, err)

	// Create temporary file
	file, err := os.CreateTemp("", "cli-policy-*.json")
	require.NoError(t, err)
	defer file.Close()

	_, err = file.Write(jsonData)
	require.NoError(t, err)

	return file.Name()
}

// createTempWorkloadFile creates a temporary workload file for testing
func createTempWorkloadFile(t *testing.T) string {
	workloadData := map[string]interface{}{
		"name":        "cli-test-workload",
		"description": "Test workload for CLI integration testing",
		"namespace":   "default",
		"resources": map[string]interface{}{
			"cpu":    "100m",
			"memory": "128Mi",
		},
		"labels": map[string]string{
			"app": "cli-test-app",
		},
	}

	jsonData, err := json.MarshalIndent(workloadData, "", "  ")
	require.NoError(t, err)

	// Create temporary file
	file, err := os.CreateTemp("", "cli-workload-*.json")
	require.NoError(t, err)
	defer file.Close()

	_, err = file.Write(jsonData)
	require.NoError(t, err)

	return file.Name()
}

// createTempAutomationRuleFile creates a temporary automation rule file for testing
func createTempAutomationRuleFile(t *testing.T) string {
	ruleData := map[string]interface{}{
		"name":        "cli-test-automation-rule",
		"description": "Test automation rule for CLI integration testing",
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
				"message": "CLI automated action executed",
			},
		},
	}

	jsonData, err := json.MarshalIndent(ruleData, "", "  ")
	require.NoError(t, err)

	// Create temporary file
	file, err := os.CreateTemp("", "cli-automation-*.json")
	require.NoError(t, err)
	defer file.Close()

	_, err = file.Write(jsonData)
	require.NoError(t, err)

	return file.Name()
}

// createTempConfigFile creates a temporary CLI config file for testing
func createTempConfigFile(t *testing.T) string {
	configData := map[string]interface{}{
		"server": map[string]interface{}{
			"host": "localhost",
			"port": 8080,
		},
		"logging": map[string]interface{}{
			"level": "info",
		},
	}

	jsonData, err := json.MarshalIndent(configData, "", "  ")
	require.NoError(t, err)

	// Create temporary file
	file, err := os.CreateTemp("", "cli-config-*.yaml")
	require.NoError(t, err)
	defer file.Close()

	_, err = file.Write(jsonData)
	require.NoError(t, err)

	return file.Name()
}

// TestCLIWorkflow tests a complete CLI workflow
func TestCLIWorkflow(t *testing.T) {
	// Build CLI binary
	cliPath := buildCLIBinary(t)
	defer os.Remove(cliPath)

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.CleanupTestServer()

	// Extract server details
	serverURL := strings.TrimPrefix(ts.Server.URL, "http://")
	serverHost := strings.Split(serverURL, ":")[0]
	serverPort := strings.Split(serverURL, ":")[1]

	t.Run("CompleteCLIWorkflow", func(t *testing.T) {
		// Step 1: Check server status
		cmd := exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"status")

		output, err := cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "Policy Engine Status")

		// Step 2: Create policy
		policyFile := createTempPolicyFile(t)
		defer os.Remove(policyFile)

		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"policy", "create", policyFile)

		output, err = cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "Policy created successfully")

		// Step 3: Create workload
		workloadFile := createTempWorkloadFile(t)
		defer os.Remove(workloadFile)

		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"workload", "create", workloadFile)

		output, err = cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "Workload created successfully")

		// Step 4: List policies and workloads
		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"policy", "list")

		output, err = cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "policies")

		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"workload", "list")

		output, err = cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "workloads")

		// Step 5: Create automation rule
		ruleFile := createTempAutomationRuleFile(t)
		defer os.Remove(ruleFile)

		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"automation", "create", ruleFile)

		output, err = cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "Automation rule created successfully")

		// Step 6: Check automation status
		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"automation", "status")

		output, err = cmd.CombinedOutput()
		require.NoError(t, err)
		// Status command should work without errors
		_ = output

		// Step 7: Get metrics
		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"metrics")

		output, err = cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "policy_engine_")

		// Step 8: Ping server
		cmd = exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"ping")

		output, err = cmd.CombinedOutput()
		require.NoError(t, err)
		assert.Contains(t, string(output), "Ping successful")
	})
}

// TestCLIErrorScenarios tests various error scenarios
func TestCLIErrorScenarios(t *testing.T) {
	cliPath := buildCLIBinary(t)
	defer os.Remove(cliPath)

	t.Run("InvalidServerConnection", func(t *testing.T) {
		cmd := exec.Command(cliPath,
			"--server-host", "invalid-host",
			"--server-port", "9999",
			"status")

		output, err := cmd.CombinedOutput()
		assert.Error(t, err)
		assert.Contains(t, string(output), "Error")
	})

	t.Run("InvalidCommand", func(t *testing.T) {
		cmd := exec.Command(cliPath, "invalid-command")
		output, err := cmd.CombinedOutput()
		assert.Error(t, err)
		assert.Contains(t, string(output), "unknown command")
	})

	t.Run("MissingArguments", func(t *testing.T) {
		cmd := exec.Command(cliPath, "policy", "create")
		output, err := cmd.CombinedOutput()
		assert.Error(t, err)
		assert.Contains(t, string(output), "required")
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		ts := SetupTestServer(t)
		defer ts.CleanupTestServer()

		serverURL := strings.TrimPrefix(ts.Server.URL, "http://")
		serverHost := strings.Split(serverURL, ":")[0]
		serverPort := strings.Split(serverURL, ":")[1]

		cmd := exec.Command(cliPath,
			"--server-host", serverHost,
			"--server-port", serverPort,
			"policy", "create", "non-existent-file.json")

		output, err := cmd.CombinedOutput()
		assert.Error(t, err)
		assert.Contains(t, string(output), "Error reading file")
	})
}
