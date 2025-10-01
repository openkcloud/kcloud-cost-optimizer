package validator

import (
	"testing"
	"time"

	"github.com/kcloud-opt/policy/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidatePolicy(t *testing.T) {
	validator := NewValidator(nil)

	t.Run("valid policy", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:      "test-policy",
				Type:      types.PolicyTypeCostOptimization,
				Status:    types.PolicyStatusActive,
				Priority:  100,
				Namespace: "default",
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
				Objectives: []types.Objective{
					{
						Type:   "cost-reduction",
						Weight: 0.5,
						Target: "20%",
					},
					{
						Type:   "performance",
						Weight: 0.5,
						Target: "95%",
					},
				},
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.NoError(t, err)
	})

	t.Run("nil policy", func(t *testing.T) {
		err := validator.ValidatePolicy(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "policy cannot be nil")
	})

	t.Run("invalid name", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name: "", // Empty name
				Type: types.PolicyTypeCostOptimization,
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("invalid name format", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name: "INVALID_NAME", // Invalid format
				Type: types.PolicyTypeCostOptimization,
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name must be a valid DNS subdomain name")
	})

	t.Run("invalid namespace", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:      "test-policy",
				Type:      types.PolicyTypeCostOptimization,
				Namespace: "INVALID_NAMESPACE",
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "namespace must be a valid DNS label name")
	})

	t.Run("invalid label key", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:   "test-policy",
				Type:   types.PolicyTypeCostOptimization,
				Labels: map[string]string{"": "value"}, // Empty key
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "label key cannot be empty")
	})

	t.Run("invalid label value", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:   "test-policy",
				Type:   types.PolicyTypeCostOptimization,
				Labels: map[string]string{"key": "INVALID_VALUE!"}, // Invalid value
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "label value must be a valid label value format")
	})

	t.Run("invalid annotation value", func(t *testing.T) {
		// Create a very long annotation value
		longValue := string(make([]byte, 262145)) // Exceeds 256KB limit
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:        "test-policy",
				Type:        types.PolicyTypeCostOptimization,
				Annotations: map[string]string{"key": longValue},
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "annotation value cannot exceed 262144 characters")
	})

	t.Run("unknown policy type", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name: "test-policy",
				Type: "unknown-type",
			},
			Spec: &types.PolicySpec{
				Type: "unknown-type",
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown policy type")
	})

	t.Run("invalid objectives", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name: "test-policy",
				Type: types.PolicyTypeCostOptimization,
			},
			Spec: &types.PolicySpec{
				Type:       types.PolicyTypeCostOptimization,
				Objectives: []types.Objective{}, // Empty objectives
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one objective is required")
	})

	t.Run("invalid objective weight", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name: "test-policy",
				Type: types.PolicyTypeCostOptimization,
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
				Objectives: []types.Objective{
					{
						Type:   "cost-reduction",
						Weight: 2.0, // Invalid weight > 1
						Target: "20%",
					},
				},
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "objective weight must be between 0 and 1")
	})

	t.Run("invalid total weight", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name: "test-policy",
				Type: types.PolicyTypeCostOptimization,
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
				Objectives: []types.Objective{
					{
						Type:   "cost-reduction",
						Weight: 0.3,
						Target: "20%",
					},
					{
						Type:   "performance",
						Weight: 0.3,
						Target: "95%",
					},
				},
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "total weight should be approximately 1.0")
	})

	t.Run("invalid rule", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name: "test-policy",
				Type: types.PolicyTypeCostOptimization,
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
				Objectives: []types.Objective{
					{
						Type:   "cost-reduction",
						Weight: 1.0,
						Target: "20%",
					},
				},
				Rules: []types.Rule{
					{
						Name:      "", // Empty name
						Condition: "workload.cpu.usage > 80%",
						Action:    "scale-up",
					},
				},
			},
		}

		err := validator.ValidatePolicy(policy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rule name cannot be empty")
	})
}

func TestValidator_ValidateWorkload(t *testing.T) {
	validator := NewValidator(nil)

	t.Run("valid workload", func(t *testing.T) {
		workload := &types.Workload{
			ID:        "workload-1",
			Name:      "test-workload",
			Type:      types.WorkloadTypeDeployment,
			Status:    types.WorkloadStatusRunning,
			Labels:    map[string]string{"app": "test"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := validator.ValidateWorkload(workload)
		assert.NoError(t, err)
	})

	t.Run("nil workload", func(t *testing.T) {
		err := validator.ValidateWorkload(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workload cannot be nil")
	})

	t.Run("empty ID", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "", // Empty ID
			Name: "test-workload",
			Type: types.WorkloadTypeDeployment,
		}

		err := validator.ValidateWorkload(workload)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workload ID cannot be empty")
	})

	t.Run("empty name", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "workload-1",
			Name: "", // Empty name
			Type: types.WorkloadTypeDeployment,
		}

		err := validator.ValidateWorkload(workload)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workload name cannot be empty")
	})

	t.Run("empty type", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "workload-1",
			Name: "test-workload",
			Type: "", // Empty type
		}

		err := validator.ValidateWorkload(workload)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workload type cannot be empty")
	})

	t.Run("empty status", func(t *testing.T) {
		workload := &types.Workload{
			ID:     "workload-1",
			Name:   "test-workload",
			Type:   types.WorkloadTypeDeployment,
			Status: "", // Empty status
		}

		err := validator.ValidateWorkload(workload)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workload status cannot be empty")
	})

	t.Run("invalid labels", func(t *testing.T) {
		workload := &types.Workload{
			ID:     "workload-1",
			Name:   "test-workload",
			Type:   types.WorkloadTypeDeployment,
			Status: types.WorkloadStatusRunning,
			Labels: map[string]string{"": "value"}, // Invalid label key
		}

		err := validator.ValidateWorkload(workload)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workload labels validation failed")
	})
}

func TestValidator_ValidateExpression(t *testing.T) {
	validator := NewValidator(nil)

	t.Run("valid expression", func(t *testing.T) {
		expression := "workload.cpu.usage > 80%"
		err := validator.ValidateExpression(expression)
		assert.NoError(t, err)
	})

	t.Run("empty expression", func(t *testing.T) {
		err := validator.ValidateExpression("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expression cannot be empty")
	})

	t.Run("expression without workload reference", func(t *testing.T) {
		expression := "some.other.variable > 80%"
		err := validator.ValidateExpression(expression)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expression must reference workload or policy")
	})

	t.Run("unbalanced parentheses", func(t *testing.T) {
		expression := "workload.cpu.usage > (80%"
		err := validator.ValidateExpression(expression)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expression has unbalanced parentheses")
	})
}

func TestValidator_ValidateTimeRange(t *testing.T) {
	validator := NewValidator(nil)

	t.Run("valid time range", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		endTime := time.Now()

		err := validator.ValidateTimeRange(startTime, endTime)
		assert.NoError(t, err)
	})

	t.Run("zero start time", func(t *testing.T) {
		endTime := time.Now()

		err := validator.ValidateTimeRange(time.Time{}, endTime)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start time cannot be zero")
	})

	t.Run("zero end time", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)

		err := validator.ValidateTimeRange(startTime, time.Time{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "end time cannot be zero")
	})

	t.Run("start time after end time", func(t *testing.T) {
		startTime := time.Now()
		endTime := time.Now().Add(-1 * time.Hour)

		err := validator.ValidateTimeRange(startTime, endTime)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start time cannot be after end time")
	})
}

func TestValidator_ValidatePercentage(t *testing.T) {
	validator := NewValidator(nil)

	t.Run("valid percentage", func(t *testing.T) {
		err := validator.ValidatePercentage("80%")
		assert.NoError(t, err)
	})

	t.Run("valid decimal percentage", func(t *testing.T) {
		err := validator.ValidatePercentage("80.5%")
		assert.NoError(t, err)
	})

	t.Run("empty percentage", func(t *testing.T) {
		err := validator.ValidatePercentage("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "percentage value cannot be empty")
	})

	t.Run("percentage without %", func(t *testing.T) {
		err := validator.ValidatePercentage("80")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "percentage value must end with %")
	})

	t.Run("percentage with empty numeric part", func(t *testing.T) {
		err := validator.ValidatePercentage("%")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "percentage value must contain a numeric part")
	})

	t.Run("percentage with invalid numeric part", func(t *testing.T) {
		err := validator.ValidatePercentage("abc%")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "percentage value must be a valid number")
	})
}

func TestValidator_ValidateAutomationRule(t *testing.T) {
	validator := NewValidator(nil)

	t.Run("valid automation rule", func(t *testing.T) {
		rule := &types.AutomationRule{
			ID:     "rule-1",
			Name:   "test-rule",
			Type:   "scale-rule",
			Status: "active",
		}

		err := validator.ValidateAutomationRule(rule)
		assert.NoError(t, err)
	})

	t.Run("nil automation rule", func(t *testing.T) {
		err := validator.ValidateAutomationRule(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "automation rule cannot be nil")
	})

	t.Run("empty ID", func(t *testing.T) {
		rule := &types.AutomationRule{
			ID:   "", // Empty ID
			Name: "test-rule",
			Type: "scale-rule",
		}

		err := validator.ValidateAutomationRule(rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "automation rule ID cannot be empty")
	})

	t.Run("empty name", func(t *testing.T) {
		rule := &types.AutomationRule{
			ID:   "rule-1",
			Name: "", // Empty name
			Type: "scale-rule",
		}

		err := validator.ValidateAutomationRule(rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "automation rule name cannot be empty")
	})

	t.Run("empty type", func(t *testing.T) {
		rule := &types.AutomationRule{
			ID:   "rule-1",
			Name: "test-rule",
			Type: "", // Empty type
		}

		err := validator.ValidateAutomationRule(rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "automation rule type cannot be empty")
	})

	t.Run("empty status", func(t *testing.T) {
		rule := &types.AutomationRule{
			ID:     "rule-1",
			Name:   "test-rule",
			Type:   "scale-rule",
			Status: "", // Empty status
		}

		err := validator.ValidateAutomationRule(rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "automation rule status cannot be empty")
	})
}
