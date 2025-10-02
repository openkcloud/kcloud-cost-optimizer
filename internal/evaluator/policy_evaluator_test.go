package evaluator

import (
	"context"
	"testing"
	"time"

	"github.com/kcloud-opt/policy/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyEvaluator_Evaluate(t *testing.T) {
	evaluator := NewPolicyEvaluator(nil, nil, nil)

	t.Run("valid policy evaluation", func(t *testing.T) {
		workload := &types.Workload{
			ID:     "workload-1",
			Name:   "test-workload",
			Type:   types.WorkloadTypeDeployment,
			Status: types.WorkloadStatusRunning,
			Labels: map[string]string{
				"app":         "test-app",
				"environment": "production",
			},
			Requirements: &types.Requirements{
				CPU:    "100m",
				Memory: "128Mi",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:      "cost-optimization-policy",
				Type:      types.PolicyTypeCostOptimization,
				Status:    types.PolicyStatusActive,
				Priority:  100,
				Namespace: "default",
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
				Target: &types.PolicyTarget{
					Namespaces: []string{"default"},
					WorkloadTypes: []types.WorkloadType{
						types.WorkloadTypeDeployment,
					},
					LabelSelectors: &types.LabelSelector{
						MatchLabels: map[string]string{
							"environment": "production",
						},
					},
				},
				Objectives: []types.Objective{
					{
						Type:   "cost-reduction",
						Weight: 1.0,
						Target: "20%",
					},
				},
				Rules: []types.Rule{
					{
						Name:      "cpu-optimization",
						Condition: "workload.cpu.usage < 50%",
						Action:    "scale-down",
					},
				},
			},
		}

		result, err := evaluator.Evaluate(context.Background(), workload, policy)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, policy.Name, result.PolicyName)
		assert.Equal(t, workload.ID, result.WorkloadID)
	})

	t.Run("nil workload", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:   "test-policy",
				Type:   types.PolicyTypeCostOptimization,
				Status: types.PolicyStatusActive,
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		}

		result, err := evaluator.Evaluate(context.Background(), nil, policy)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("nil policy", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "workload-1",
			Name: "test-workload",
			Type: types.WorkloadTypeDeployment,
		}

		result, err := evaluator.Evaluate(context.Background(), workload, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("inactive policy", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "workload-1",
			Name: "test-workload",
			Type: types.WorkloadTypeDeployment,
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:   "test-policy",
				Type:   types.PolicyTypeCostOptimization,
				Status: types.PolicyStatusInactive, // Inactive policy
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		}

		result, err := evaluator.Evaluate(context.Background(), workload, policy)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "policy is not active")
	})

	t.Run("policy not applicable to workload", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "workload-1",
			Name: "test-workload",
			Type: types.WorkloadTypeDeployment,
			Labels: map[string]string{
				"environment": "development", // Different environment
			},
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:   "test-policy",
				Type:   types.PolicyTypeCostOptimization,
				Status: types.PolicyStatusActive,
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
				Target: &types.PolicyTarget{
					LabelSelectors: &types.LabelSelector{
						MatchLabels: map[string]string{
							"environment": "production", // Different environment
						},
					},
				},
			},
		}

		result, err := evaluator.Evaluate(context.Background(), workload, policy)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "policy not applicable")
	})
}

func TestPolicyEvaluator_EvaluateSingle(t *testing.T) {
	evaluator := NewPolicyEvaluator(nil, nil, nil)

	t.Run("evaluate single policy", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "workload-1",
			Name: "test-workload",
			Type: types.WorkloadTypeDeployment,
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:   "test-policy",
				Type:   types.PolicyTypeCostOptimization,
				Status: types.PolicyStatusActive,
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		}

		result, err := evaluator.EvaluateSingle(context.Background(), workload, policy)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, policy.Name, result.PolicyName)
		assert.Equal(t, workload.ID, result.WorkloadID)
	})

	t.Run("evaluate with invalid policy", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "workload-1",
			Name: "test-workload",
			Type: types.WorkloadTypeDeployment,
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:   "test-policy",
				Type:   "invalid-type", // Invalid policy type
				Status: types.PolicyStatusActive,
			},
			Spec: &types.PolicySpec{
				Type: "invalid-type",
			},
		}

		result, err := evaluator.EvaluateSingle(context.Background(), workload, policy)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestPolicyEvaluator_GetApplicablePolicies(t *testing.T) {
	evaluator := NewPolicyEvaluator(nil, nil, nil)

	t.Run("get applicable policies", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "workload-1",
			Name: "test-workload",
			Type: types.WorkloadTypeDeployment,
			Labels: map[string]string{
				"environment": "production",
				"app":         "test-app",
			},
		}

		policies := []types.Policy{
			{
				Metadata: types.PolicyMetadata{
					Name:   "policy-1",
					Type:   types.PolicyTypeCostOptimization,
					Status: types.PolicyStatusActive,
				},
				Spec: &types.PolicySpec{
					Type: types.PolicyTypeCostOptimization,
					Target: &types.PolicyTarget{
						LabelSelectors: &types.LabelSelector{
							MatchLabels: map[string]string{
								"environment": "production",
							},
						},
					},
				},
			},
			{
				Metadata: types.PolicyMetadata{
					Name:   "policy-2",
					Type:   types.PolicyTypeAutomation,
					Status: types.PolicyStatusActive,
				},
				Spec: &types.PolicySpec{
					Type: types.PolicyTypeAutomation,
					Target: &types.PolicyTarget{
						LabelSelectors: &types.LabelSelector{
							MatchLabels: map[string]string{
								"environment": "development", // Different environment
							},
						},
					},
				},
			},
			{
				Metadata: types.PolicyMetadata{
					Name:   "policy-3",
					Type:   types.PolicyTypeSecurity,
					Status: types.PolicyStatusInactive, // Inactive policy
				},
				Spec: &types.PolicySpec{
					Type: types.PolicyTypeSecurity,
				},
			},
		}

		applicablePolicies := evaluator.GetApplicablePolicies(workload, policies)
		assert.Len(t, applicablePolicies, 1)
		assert.Equal(t, "policy-1", applicablePolicies[0].Name)
	})

	t.Run("no applicable policies", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "workload-1",
			Name: "test-workload",
			Type: types.WorkloadTypeDeployment,
			Labels: map[string]string{
				"environment": "development",
			},
		}

		policies := []types.Policy{
			{
				Metadata: types.PolicyMetadata{
					Name:   "policy-1",
					Type:   types.PolicyTypeCostOptimization,
					Status: types.PolicyStatusActive,
				},
				Spec: &types.PolicySpec{
					Type: types.PolicyTypeCostOptimization,
					Target: &types.PolicyTarget{
						LabelSelectors: &types.LabelSelector{
							MatchLabels: map[string]string{
								"environment": "production", // Different environment
							},
						},
					},
				},
			},
		}

		applicablePolicies := evaluator.GetApplicablePolicies(workload, policies)
		assert.Len(t, applicablePolicies, 0)
	})
}

func TestPolicyEvaluator_ValidatePolicy(t *testing.T) {
	evaluator := NewPolicyEvaluator(nil, nil, nil)

	t.Run("valid policy", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:   "test-policy",
				Type:   types.PolicyTypeCostOptimization,
				Status: types.PolicyStatusActive,
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
			},
		}

		err := evaluator.ValidatePolicy(policy)
		assert.NoError(t, err)
	})

	t.Run("invalid policy", func(t *testing.T) {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:   "test-policy",
				Type:   types.PolicyTypeCostOptimization,
				Status: types.PolicyStatusActive,
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
				Objectives: []types.Objective{
					{
						Type:   "cost-reduction",
						Weight: 2.0, // Invalid weight
						Target: "20%",
					},
				},
			},
		}

		err := evaluator.ValidatePolicy(policy)
		assert.Error(t, err)
	})
}

func TestPolicyEvaluator_Health(t *testing.T) {
	evaluator := NewPolicyEvaluator(nil, nil, nil)

	health, err := evaluator.Health(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "healthy", health["status"])
}

func TestPolicyEvaluator_isPolicyApplicable(t *testing.T) {
	evaluator := NewPolicyEvaluator(nil, nil, nil)

	t.Run("policy applicable by namespace", func(t *testing.T) {
		workload := &types.Workload{
			ID:        "workload-1",
			Namespace: "default",
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Status: types.PolicyStatusActive,
			},
			Spec: &types.PolicySpec{
				Target: &types.PolicyTarget{
					Namespaces: []string{"default"},
				},
			},
		}

		applicable := evaluator.isPolicyApplicable(workload, policy)
		assert.True(t, applicable)
	})

	t.Run("policy not applicable by namespace", func(t *testing.T) {
		workload := &types.Workload{
			ID:        "workload-1",
			Namespace: "default",
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Status: types.PolicyStatusActive,
			},
			Spec: &types.PolicySpec{
				Target: &types.PolicyTarget{
					Namespaces: []string{"production"}, // Different namespace
				},
			},
		}

		applicable := evaluator.isPolicyApplicable(workload, policy)
		assert.False(t, applicable)
	})

	t.Run("policy applicable by workload type", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "workload-1",
			Type: types.WorkloadTypeDeployment,
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Status: types.PolicyStatusActive,
			},
			Spec: &types.PolicySpec{
				Target: &types.PolicyTarget{
					WorkloadTypes: []types.WorkloadType{
						types.WorkloadTypeDeployment,
					},
				},
			},
		}

		applicable := evaluator.isPolicyApplicable(workload, policy)
		assert.True(t, applicable)
	})

	t.Run("policy not applicable by workload type", func(t *testing.T) {
		workload := &types.Workload{
			ID:   "workload-1",
			Type: types.WorkloadTypeDeployment,
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Status: types.PolicyStatusActive,
			},
			Spec: &types.PolicySpec{
				Target: &types.PolicyTarget{
					WorkloadTypes: []types.WorkloadType{
						types.WorkloadTypeStatefulSet, // Different type
					},
				},
			},
		}

		applicable := evaluator.isPolicyApplicable(workload, policy)
		assert.False(t, applicable)
	})

	t.Run("policy applicable by labels", func(t *testing.T) {
		workload := &types.Workload{
			ID: "workload-1",
			Labels: map[string]string{
				"environment": "production",
				"app":         "test-app",
			},
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Status: types.PolicyStatusActive,
			},
			Spec: &types.PolicySpec{
				Target: &types.PolicyTarget{
					LabelSelectors: &types.LabelSelector{
						MatchLabels: map[string]string{
							"environment": "production",
						},
					},
				},
			},
		}

		applicable := evaluator.isPolicyApplicable(workload, policy)
		assert.True(t, applicable)
	})

	t.Run("policy not applicable by labels", func(t *testing.T) {
		workload := &types.Workload{
			ID: "workload-1",
			Labels: map[string]string{
				"environment": "development",
			},
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Status: types.PolicyStatusActive,
			},
			Spec: &types.PolicySpec{
				Target: &types.PolicyTarget{
					LabelSelectors: &types.LabelSelector{
						MatchLabels: map[string]string{
							"environment": "production", // Different environment
						},
					},
				},
			},
		}

		applicable := evaluator.isPolicyApplicable(workload, policy)
		assert.False(t, applicable)
	})

	t.Run("inactive policy", func(t *testing.T) {
		workload := &types.Workload{
			ID: "workload-1",
		}

		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Status: types.PolicyStatusInactive, // Inactive policy
			},
			Spec: &types.PolicySpec{},
		}

		applicable := evaluator.isPolicyApplicable(workload, policy)
		assert.False(t, applicable)
	})
}

func TestPolicyEvaluator_calculateCostScore(t *testing.T) {
	evaluator := NewPolicyEvaluator(nil, nil, nil)

	t.Run("calculate cost score", func(t *testing.T) {
		workload := &types.Workload{
			ID: "workload-1",
			Requirements: &types.Requirements{
				CPU:    "500m",
				Memory: "1Gi",
			},
		}

		policy := &types.Policy{
			Spec: &types.PolicySpec{
				Objectives: []types.Objective{
					{
						Type:   "cost-reduction",
						Weight: 1.0,
						Target: "20%",
					},
				},
			},
		}

		score := evaluator.calculateCostScore(workload, policy)
		assert.GreaterOrEqual(t, score, 0.0)
		assert.LessOrEqual(t, score, 100.0)
	})
}

func TestPolicyEvaluator_calculatePriorityScore(t *testing.T) {
	evaluator := NewPolicyEvaluator(nil, nil, nil)

	t.Run("calculate priority score", func(t *testing.T) {
		workload := &types.Workload{
			ID: "workload-1",
			Labels: map[string]string{
				"priority": "high",
			},
		}

		policy := &types.Policy{
			Spec: &types.PolicySpec{
				Objectives: []types.Objective{
					{
						Type:   "priority",
						Weight: 1.0,
						Target: "high",
					},
				},
			},
		}

		score := evaluator.calculatePriorityScore(workload, policy)
		assert.GreaterOrEqual(t, score, 0.0)
		assert.LessOrEqual(t, score, 100.0)
	})
}
