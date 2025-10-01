package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicy_Validate(t *testing.T) {
	t.Run("valid policy", func(t *testing.T) {
		policy := &Policy{
			Metadata: PolicyMetadata{
				Name:      "test-policy",
				Type:      PolicyTypeCostOptimization,
				Status:    PolicyStatusActive,
				Priority:  100,
				Namespace: "default",
			},
			Spec: &PolicySpec{
				Type: PolicyTypeCostOptimization,
				Objectives: []Objective{
					{
						Type:   "cost-reduction",
						Weight: 1.0,
						Target: "20%",
					},
				},
			},
		}

		err := policy.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid name", func(t *testing.T) {
		policy := &Policy{
			Metadata: PolicyMetadata{
				Name: "", // Empty name
				Type: PolicyTypeCostOptimization,
			},
			Spec: &PolicySpec{
				Type: PolicyTypeCostOptimization,
			},
		}

		err := policy.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("invalid type", func(t *testing.T) {
		policy := &Policy{
			Metadata: PolicyMetadata{
				Name: "test-policy",
				Type: "invalid-type",
			},
			Spec: &PolicySpec{
				Type: "invalid-type",
			},
		}

		err := policy.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type cannot be empty")
	})

	t.Run("nil spec", func(t *testing.T) {
		policy := &Policy{
			Metadata: PolicyMetadata{
				Name: "test-policy",
				Type: PolicyTypeCostOptimization,
			},
			Spec: nil, // Nil spec
		}

		err := policy.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "spec cannot be nil")
	})
}

func TestPolicy_GetMetadata(t *testing.T) {
	policy := &Policy{
		Metadata: PolicyMetadata{
			Name:      "test-policy",
			Type:      PolicyTypeCostOptimization,
			Status:    PolicyStatusActive,
			Priority:  100,
			Namespace: "default",
		},
	}

	metadata := policy.GetMetadata()
	assert.Equal(t, "test-policy", metadata.Name)
	assert.Equal(t, PolicyTypeCostOptimization, metadata.Type)
	assert.Equal(t, PolicyStatusActive, metadata.Status)
	assert.Equal(t, 100, metadata.Priority)
	assert.Equal(t, "default", metadata.Namespace)
}

func TestPolicy_SetStatus(t *testing.T) {
	policy := &Policy{
		Metadata: PolicyMetadata{
			Name:   "test-policy",
			Status: PolicyStatusActive,
		},
	}

	policy.SetStatus(PolicyStatusInactive)
	assert.Equal(t, PolicyStatusInactive, policy.Status)
}

func TestPolicy_SetPriority(t *testing.T) {
	policy := &Policy{
		Metadata: PolicyMetadata{
			Name:     "test-policy",
			Priority: 100,
		},
	}

	policy.SetPriority(200)
	assert.Equal(t, 200, policy.Priority)
}

func TestWorkload_Validate(t *testing.T) {
	t.Run("valid workload", func(t *testing.T) {
		workload := &Workload{
			ID:        "workload-1",
			Name:      "test-workload",
			Type:      WorkloadTypeDeployment,
			Status:    WorkloadStatusRunning,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := workload.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		workload := &Workload{
			ID:   "", // Empty ID
			Name: "test-workload",
			Type: WorkloadTypeDeployment,
		}

		err := workload.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID cannot be empty")
	})

	t.Run("empty name", func(t *testing.T) {
		workload := &Workload{
			ID:   "workload-1",
			Name: "", // Empty name
			Type: WorkloadTypeDeployment,
		}

		err := workload.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("empty type", func(t *testing.T) {
		workload := &Workload{
			ID:   "workload-1",
			Name: "test-workload",
			Type: "", // Empty type
		}

		err := workload.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type cannot be empty")
	})

	t.Run("empty status", func(t *testing.T) {
		workload := &Workload{
			ID:     "workload-1",
			Name:   "test-workload",
			Type:   WorkloadTypeDeployment,
			Status: "", // Empty status
		}

		err := workload.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status cannot be empty")
	})
}

func TestWorkload_ParseMemory(t *testing.T) {
	t.Run("parse memory string", func(t *testing.T) {
		workload := &Workload{}

		// Test various memory formats
		testCases := []struct {
			input    string
			expected int64
		}{
			{"128Mi", 128 * 1024 * 1024},
			{"1Gi", 1024 * 1024 * 1024},
			{"2Gi", 2 * 1024 * 1024 * 1024},
			{"512Mi", 512 * 1024 * 1024},
		}

		for _, tc := range testCases {
			result, err := workload.ParseMemory(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		}
	})

	t.Run("invalid memory format", func(t *testing.T) {
		workload := &Workload{}

		_, err := workload.ParseMemory("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid memory format")
	})
}

func TestDecision_Validate(t *testing.T) {
	t.Run("valid decision", func(t *testing.T) {
		decision := &Decision{
			ID:           "decision-1",
			WorkloadID:   "workload-1",
			PolicyID:     "policy-1",
			DecisionType: DecisionTypeScaleUp,
			Status:       DecisionStatusPending,
			CreatedAt:    time.Now(),
		}

		err := decision.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		decision := &Decision{
			ID: "", // Empty ID
		}

		err := decision.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID cannot be empty")
	})

	t.Run("empty workload ID", func(t *testing.T) {
		decision := &Decision{
			ID:         "decision-1",
			WorkloadID: "", // Empty workload ID
		}

		err := decision.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "WorkloadID cannot be empty")
	})

	t.Run("empty decision type", func(t *testing.T) {
		decision := &Decision{
			ID:           "decision-1",
			WorkloadID:   "workload-1",
			DecisionType: "", // Empty decision type
		}

		err := decision.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DecisionType cannot be empty")
	})

	t.Run("empty status", func(t *testing.T) {
		decision := &Decision{
			ID:         "decision-1",
			WorkloadID: "workload-1",
			Status:     "", // Empty status
		}

		err := decision.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Status cannot be empty")
	})
}

func TestEvaluation_Validate(t *testing.T) {
	t.Run("valid evaluation", func(t *testing.T) {
		evaluation := &Evaluation{
			ID:             "evaluation-1",
			WorkloadID:     "workload-1",
			PolicyID:       "policy-1",
			EvaluationType: EvaluationTypeCostOptimization,
			Status:         EvaluationStatusCompleted,
			Result:         EvaluationResultPass,
			CreatedAt:      time.Now(),
		}

		err := evaluation.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		evaluation := &Evaluation{
			ID: "", // Empty ID
		}

		err := evaluation.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID cannot be empty")
	})

	t.Run("empty workload ID", func(t *testing.T) {
		evaluation := &Evaluation{
			ID:         "evaluation-1",
			WorkloadID: "", // Empty workload ID
		}

		err := evaluation.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "WorkloadID cannot be empty")
	})

	t.Run("empty evaluation type", func(t *testing.T) {
		evaluation := &Evaluation{
			ID:             "evaluation-1",
			WorkloadID:     "workload-1",
			EvaluationType: "", // Empty evaluation type
		}

		err := evaluation.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EvaluationType cannot be empty")
	})

	t.Run("empty status", func(t *testing.T) {
		evaluation := &Evaluation{
			ID:         "evaluation-1",
			WorkloadID: "workload-1",
			Status:     "", // Empty status
		}

		err := evaluation.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Status cannot be empty")
	})
}

func TestAutomationRule_Validate(t *testing.T) {
	t.Run("valid automation rule", func(t *testing.T) {
		rule := &AutomationRule{
			ID:     "rule-1",
			Name:   "test-rule",
			Type:   "scale-rule",
			Status: "active",
		}

		err := rule.Validate()
		assert.NoError(t, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		rule := &AutomationRule{
			ID: "", // Empty ID
		}

		err := rule.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID cannot be empty")
	})

	t.Run("empty name", func(t *testing.T) {
		rule := &AutomationRule{
			ID:   "rule-1",
			Name: "", // Empty name
		}

		err := rule.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("empty type", func(t *testing.T) {
		rule := &AutomationRule{
			ID:   "rule-1",
			Name: "test-rule",
			Type: "", // Empty type
		}

		err := rule.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type cannot be empty")
	})

	t.Run("empty status", func(t *testing.T) {
		rule := &AutomationRule{
			ID:     "rule-1",
			Name:   "test-rule",
			Type:   "scale-rule",
			Status: "", // Empty status
		}

		err := rule.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status cannot be empty")
	})
}

func TestPolicyConstants(t *testing.T) {
	// Test policy type constants
	assert.Equal(t, "cost-optimization", string(PolicyTypeCostOptimization))
	assert.Equal(t, "automation", string(PolicyTypeAutomation))
	assert.Equal(t, "workload-priority", string(PolicyTypeWorkloadPriority))
	assert.Equal(t, "security", string(PolicyTypeSecurity))
	assert.Equal(t, "resource-quota", string(PolicyTypeResourceQuota))

	// Test policy status constants
	assert.Equal(t, "active", string(PolicyStatusActive))
	assert.Equal(t, "inactive", string(PolicyStatusInactive))
	assert.Equal(t, "draft", string(PolicyStatusDraft))

	// Test workload type constants
	assert.Equal(t, "deployment", string(WorkloadTypeDeployment))
	assert.Equal(t, "statefulset", string(WorkloadTypeStatefulSet))
	assert.Equal(t, "daemonset", string(WorkloadTypeDaemonSet))
	assert.Equal(t, "job", string(WorkloadTypeJob))
	assert.Equal(t, "cronjob", string(WorkloadTypeCronJob))

	// Test workload status constants
	assert.Equal(t, "running", string(WorkloadStatusRunning))
	assert.Equal(t, "stopped", string(WorkloadStatusStopped))
	assert.Equal(t, "pending", string(WorkloadStatusPending))
	assert.Equal(t, "failed", string(WorkloadStatusFailed))

	// Test decision type constants
	assert.Equal(t, "scale-up", string(DecisionTypeScaleUp))
	assert.Equal(t, "scale-down", string(DecisionTypeScaleDown))
	assert.Equal(t, "resource-adjustment", string(DecisionTypeResourceAdjustment))
	assert.Equal(t, "notification", string(DecisionTypeNotification))

	// Test decision status constants
	assert.Equal(t, "pending", string(DecisionStatusPending))
	assert.Equal(t, "executed", string(DecisionStatusExecuted))
	assert.Equal(t, "failed", string(DecisionStatusFailed))
	assert.Equal(t, "cancelled", string(DecisionStatusCancelled))

	// Test evaluation type constants
	assert.Equal(t, "cost-optimization", string(EvaluationTypeCostOptimization))
	assert.Equal(t, "automation", string(EvaluationTypeAutomation))
	assert.Equal(t, "workload-priority", string(EvaluationTypeWorkloadPriority))
	assert.Equal(t, "security", string(EvaluationTypeSecurity))

	// Test evaluation status constants
	assert.Equal(t, "pending", string(EvaluationStatusPending))
	assert.Equal(t, "running", string(EvaluationStatusRunning))
	assert.Equal(t, "completed", string(EvaluationStatusCompleted))
	assert.Equal(t, "failed", string(EvaluationStatusFailed))

	// Test evaluation result constants
	assert.Equal(t, "pass", string(EvaluationResultPass))
	assert.Equal(t, "fail", string(EvaluationResultFail))
	assert.Equal(t, "warning", string(EvaluationResultWarning))
	assert.Equal(t, "error", string(EvaluationResultError))
}

func TestPriorityConstants(t *testing.T) {
	// Test priority constants
	assert.Equal(t, 1000, int(PriorityCritical))
	assert.Equal(t, 750, int(PriorityHigh))
	assert.Equal(t, 500, int(PriorityMedium))
	assert.Equal(t, 250, int(PriorityLow))
	assert.Equal(t, 100, int(PriorityMinimal))
}
