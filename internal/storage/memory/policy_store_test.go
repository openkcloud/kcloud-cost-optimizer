package memory

import (
	"context"
	"fmt"
	"testing"

	"github.com/kcloud-opt/policy/internal/storage"
	"github.com/kcloud-opt/policy/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyStore_Create(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

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
		},
	}

	err := store.Create(ctx, *policy)
	require.NoError(t, err)

	// Verify policy was created
	retrieved, err := store.Get(ctx, "test-policy")
	require.NoError(t, err)
	assert.Equal(t, policy.Name, retrieved.Name)
	assert.Equal(t, policy.Type, retrieved.Type)
}

func TestPolicyStore_CreateDuplicate(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

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
		},
	}

	err := store.Create(ctx, *policy)
	require.NoError(t, err)

	// Try to create duplicate
	err = store.Create(ctx, *policy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy already exists")
}

func TestPolicyStore_Get(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

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
		},
	}

	err := store.Create(ctx, *policy)
	require.NoError(t, err)

	retrieved, err := store.Get(ctx, "test-policy")
	require.NoError(t, err)
	assert.Equal(t, policy.Name, retrieved.Name)
	assert.Equal(t, policy.Type, retrieved.Type)
	assert.Equal(t, policy.Status, retrieved.Status)
}

func TestPolicyStore_GetNotFound(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

	_, err := store.Get(ctx, "non-existent-policy")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy not found")
}

func TestPolicyStore_Update(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

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
		},
	}

	err := store.Create(ctx, *policy)
	require.NoError(t, err)

	// Update policy
	policy.Status = types.PolicyStatusInactive
	policy.Priority = 200

	err = store.Update(ctx, *policy)
	require.NoError(t, err)

	// Verify update
	retrieved, err := store.Get(ctx, "test-policy")
	require.NoError(t, err)
	assert.Equal(t, types.PolicyStatusInactive, retrieved.Status)
	assert.Equal(t, 200, retrieved.Priority)
}

func TestPolicyStore_UpdateNotFound(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

	policy := &types.Policy{
		Metadata: types.PolicyMetadata{
			Name:      "non-existent-policy",
			Type:      types.PolicyTypeCostOptimization,
			Status:    types.PolicyStatusActive,
			Priority:  100,
			Namespace: "default",
		},
		Spec: &types.PolicySpec{
			Type: types.PolicyTypeCostOptimization,
		},
	}

	err := store.Update(ctx, *policy)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy not found")
}

func TestPolicyStore_Delete(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

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
		},
	}

	err := store.Create(ctx, *policy)
	require.NoError(t, err)

	err = store.Delete(ctx, "test-policy")
	require.NoError(t, err)

	// Verify deletion
	_, err = store.Get(ctx, "test-policy")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy not found")
}

func TestPolicyStore_DeleteNotFound(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

	err := store.Delete(ctx, "non-existent-policy")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy not found")
}

func TestPolicyStore_List(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

	// Create test policies
	policies := []types.Policy{
		{
			Metadata: types.PolicyMetadata{
				Name:      "policy-1",
				Type:      types.PolicyTypeCostOptimization,
				Status:    types.PolicyStatusActive,
				Priority:  100,
				Namespace: "default",
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		},
		{
			Metadata: types.PolicyMetadata{
				Name:      "policy-2",
				Type:      types.PolicyTypeAutomation,
				Status:    types.PolicyStatusInactive,
				Priority:  200,
				Namespace: "production",
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeAutomation,
			},
		},
	}

	for _, policy := range policies {
		err := store.Create(ctx, policy)
		require.NoError(t, err)
	}

	// Test list all
	allPolicies, err := store.List(ctx, &storage.PolicyFilters{})
	require.NoError(t, err)
	assert.Len(t, allPolicies, 2)

	// Test filter by type
	costPolicies, err := store.List(ctx, &storage.PolicyFilters{
		Type: func() *types.PolicyType { t := types.PolicyTypeCostOptimization; return &t }(),
	})
	require.NoError(t, err)
	assert.Len(t, costPolicies, 1)
	assert.Equal(t, "policy-1", costPolicies[0].Name)

	// Test filter by status
	activePolicies, err := store.List(ctx, &storage.PolicyFilters{
		Status: func() *types.PolicyStatus { s := types.PolicyStatusActive; return &s }(),
	})
	require.NoError(t, err)
	assert.Len(t, activePolicies, 1)
	assert.Equal(t, "policy-1", activePolicies[0].Name)

	// Test filter by namespace
	defaultPolicies, err := store.List(ctx, &storage.PolicyFilters{
		Namespace: func() *string { n := "default"; return &n }(),
	})
	require.NoError(t, err)
	assert.Len(t, defaultPolicies, 1)
	assert.Equal(t, "policy-1", defaultPolicies[0].Name)
}

func TestPolicyStore_Count(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

	// Create test policies
	policies := []types.Policy{
		{
			Metadata: types.PolicyMetadata{
				Name:      "policy-1",
				Type:      types.PolicyTypeCostOptimization,
				Status:    types.PolicyStatusActive,
				Priority:  100,
				Namespace: "default",
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		},
		{
			Metadata: types.PolicyMetadata{
				Name:      "policy-2",
				Type:      types.PolicyTypeAutomation,
				Status:    types.PolicyStatusInactive,
				Priority:  200,
				Namespace: "production",
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeAutomation,
			},
		},
	}

	for _, policy := range policies {
		err := store.Create(ctx, policy)
		require.NoError(t, err)
	}

	// Test count all
	count, err := store.Count(ctx, &storage.PolicyFilters{})
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Test count with filter
	activeCount, err := store.Count(ctx, &storage.PolicyFilters{
		Status: func() *types.PolicyStatus { s := types.PolicyStatusActive; return &s }(),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, activeCount)
}

func TestPolicyStore_Search(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

	// Create test policies
	policies := []types.Policy{
		{
			Metadata: types.PolicyMetadata{
				Name:      "cost-optimization-policy",
				Type:      types.PolicyTypeCostOptimization,
				Status:    types.PolicyStatusActive,
				Priority:  100,
				Namespace: "default",
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		},
		{
			Metadata: types.PolicyMetadata{
				Name:      "automation-rule",
				Type:      types.PolicyTypeAutomation,
				Status:    types.PolicyStatusInactive,
				Priority:  200,
				Namespace: "production",
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeAutomation,
			},
		},
	}

	for _, policy := range policies {
		err := store.Create(ctx, policy)
		require.NoError(t, err)
	}

	// Test search by name
	results, err := store.Search(ctx, &storage.PolicySearchQuery{
		Query: "cost",
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "cost-optimization-policy", results[0].Name)

	// Test search by type
	results, err = store.Search(ctx, &storage.PolicySearchQuery{
		Query: "automation",
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "automation-rule", results[0].Name)
}

func TestPolicyStore_GetVersions(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

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
		},
	}

	err := store.Create(ctx, *policy)
	require.NoError(t, err)

	// Update policy to create a new version
	policy.Priority = 200
	err = store.Update(ctx, *policy)
	require.NoError(t, err)

	versions, err := store.GetVersions(ctx, "test-policy")
	require.NoError(t, err)
	assert.Len(t, versions, 2)

	// Verify versions are ordered by creation time
	assert.Equal(t, 100, versions[0].Priority)
	assert.Equal(t, 200, versions[1].Priority)
}

func TestPolicyStore_Health(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

	health, err := store.Health(ctx)
	require.NoError(t, err)
	assert.Equal(t, "healthy", health["status"])
}

func TestPolicyStore_GetMetrics(t *testing.T) {
	store := NewPolicyStore()
	ctx := context.Background()

	// Create some policies
	for i := 0; i < 3; i++ {
		policy := &types.Policy{
			Metadata: types.PolicyMetadata{
				Name:      fmt.Sprintf("policy-%d", i),
				Type:      types.PolicyTypeCostOptimization,
				Status:    types.PolicyStatusActive,
				Priority:  100,
				Namespace: "default",
			},
			Spec: &types.PolicySpec{
				Type: types.PolicyTypeCostOptimization,
			},
		}
		err := store.Create(ctx, *policy)
		require.NoError(t, err)
	}

	metrics, err := store.GetMetrics(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, metrics["total_policies"])
	assert.Equal(t, 3, metrics["active_policies"])
}
