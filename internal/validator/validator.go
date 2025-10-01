package validator

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kcloud-opt/policy/internal/types"
)

// Validator provides policy validation functionality
type Validator struct {
	logger *types.Logger
}

// NewValidator creates a new validator instance
func NewValidator(logger *types.Logger) *Validator {
	return &Validator{
		logger: logger,
	}
}

// ValidatePolicy validates a policy against all validation rules
func (v *Validator) ValidatePolicy(policy *types.Policy) error {
	if policy == nil {
		return fmt.Errorf("policy cannot be nil")
	}

	// Validate metadata
	if err := v.validateMetadata(&policy.Metadata); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	// Validate spec based on policy type
	switch policy.Type {
	case types.PolicyTypeCostOptimization:
		if err := v.validateCostOptimizationPolicy(policy); err != nil {
			return fmt.Errorf("cost optimization policy validation failed: %w", err)
		}
	case types.PolicyTypeAutomation:
		if err := v.validateAutomationPolicy(policy); err != nil {
			return fmt.Errorf("automation policy validation failed: %w", err)
		}
	case types.PolicyTypeWorkloadPriority:
		if err := v.validateWorkloadPriorityPolicy(policy); err != nil {
			return fmt.Errorf("workload priority policy validation failed: %w", err)
		}
	case types.PolicyTypeSecurity:
		if err := v.validateSecurityPolicy(policy); err != nil {
			return fmt.Errorf("security policy validation failed: %w", err)
		}
	case types.PolicyTypeResourceQuota:
		if err := v.validateResourceQuotaPolicy(policy); err != nil {
			return fmt.Errorf("resource quota policy validation failed: %w", err)
		}
	default:
		return fmt.Errorf("unknown policy type: %s", policy.Type)
	}

	return nil
}

// validateMetadata validates policy metadata
func (v *Validator) validateMetadata(metadata *types.PolicyMetadata) error {
	if metadata == nil {
		return fmt.Errorf("metadata cannot be nil")
	}

	// Validate name
	if err := v.validateName(metadata.Name); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}

	// Validate namespace
	if metadata.Namespace != "" {
		if err := v.validateNamespace(metadata.Namespace); err != nil {
			return fmt.Errorf("invalid namespace: %w", err)
		}
	}

	// Validate labels
	if err := v.validateLabels(metadata.Labels); err != nil {
		return fmt.Errorf("invalid labels: %w", err)
	}

	// Validate annotations
	if err := v.validateAnnotations(metadata.Annotations); err != nil {
		return fmt.Errorf("invalid annotations: %w", err)
	}

	return nil
}

// validateName validates policy name
func (v *Validator) validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 253 {
		return fmt.Errorf("name cannot exceed 253 characters")
	}

	// Check for valid DNS subdomain name
	nameRegex := regexp.MustCompile(`^[a-z0-9]([a-z0-9\-]*[a-z0-9])?$`)
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("name must be a valid DNS subdomain name (lowercase alphanumeric and hyphens only)")
	}

	return nil
}

// validateNamespace validates namespace
func (v *Validator) validateNamespace(namespace string) error {
	if len(namespace) > 63 {
		return fmt.Errorf("namespace cannot exceed 63 characters")
	}

	namespaceRegex := regexp.MustCompile(`^[a-z0-9]([a-z0-9\-]*[a-z0-9])?$`)
	if !namespaceRegex.MatchString(namespace) {
		return fmt.Errorf("namespace must be a valid DNS label name")
	}

	return nil
}

// validateLabels validates labels
func (v *Validator) validateLabels(labels map[string]string) error {
	if labels == nil {
		return nil
	}

	for key, value := range labels {
		if err := v.validateLabelKey(key); err != nil {
			return fmt.Errorf("invalid label key %s: %w", key, err)
		}
		if err := v.validateLabelValue(value); err != nil {
			return fmt.Errorf("invalid label value for key %s: %w", key, err)
		}
	}

	return nil
}

// validateLabelKey validates label key
func (v *Validator) validateLabelKey(key string) error {
	if key == "" {
		return fmt.Errorf("label key cannot be empty")
	}

	if len(key) > 253 {
		return fmt.Errorf("label key cannot exceed 253 characters")
	}

	// Check for valid label key format
	keyRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-_.]*[a-zA-Z0-9])?/)?[a-zA-Z0-9]([a-zA-Z0-9\-_.]*[a-zA-Z0-9])?$`)
	if !keyRegex.MatchString(key) {
		return fmt.Errorf("label key must be a valid label key format")
	}

	return nil
}

// validateLabelValue validates label value
func (v *Validator) validateLabelValue(value string) error {
	if len(value) > 63 {
		return fmt.Errorf("label value cannot exceed 63 characters")
	}

	valueRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-_.]*[a-zA-Z0-9])?$`)
	if !valueRegex.MatchString(value) {
		return fmt.Errorf("label value must be a valid label value format")
	}

	return nil
}

// validateAnnotations validates annotations
func (v *Validator) validateAnnotations(annotations map[string]string) error {
	if annotations == nil {
		return nil
	}

	for key, value := range annotations {
		if err := v.validateAnnotationKey(key); err != nil {
			return fmt.Errorf("invalid annotation key %s: %w", key, err)
		}
		if err := v.validateAnnotationValue(value); err != nil {
			return fmt.Errorf("invalid annotation value for key %s: %w", key, err)
		}
	}

	return nil
}

// validateAnnotationKey validates annotation key
func (v *Validator) validateAnnotationKey(key string) error {
	if key == "" {
		return fmt.Errorf("annotation key cannot be empty")
	}

	if len(key) > 253 {
		return fmt.Errorf("annotation key cannot exceed 253 characters")
	}

	keyRegex := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-_.]*[a-zA-Z0-9])?/)?[a-zA-Z0-9]([a-zA-Z0-9\-_.]*[a-zA-Z0-9])?$`)
	if !keyRegex.MatchString(key) {
		return fmt.Errorf("annotation key must be a valid annotation key format")
	}

	return nil
}

// validateAnnotationValue validates annotation value
func (v *Validator) validateAnnotationValue(value string) error {
	if len(value) > 262144 { // 256KB limit
		return fmt.Errorf("annotation value cannot exceed 262144 characters")
	}

	return nil
}

// validateCostOptimizationPolicy validates cost optimization policy
func (v *Validator) validateCostOptimizationPolicy(policy *types.Policy) error {
	if policy.Spec == nil {
		return fmt.Errorf("spec cannot be nil")
	}

	// Validate objectives
	if err := v.validateObjectives(policy.Spec.Objectives); err != nil {
		return fmt.Errorf("objectives validation failed: %w", err)
	}

	// Validate constraints
	if err := v.validateConstraints(policy.Spec.Constraints); err != nil {
		return fmt.Errorf("constraints validation failed: %w", err)
	}

	// Validate rules
	if err := v.validateRules(policy.Spec.Rules); err != nil {
		return fmt.Errorf("rules validation failed: %w", err)
	}

	// Validate actions
	if err := v.validateActions(policy.Spec.Actions); err != nil {
		return fmt.Errorf("actions validation failed: %w", err)
	}

	return nil
}

// validateObjectives validates policy objectives
func (v *Validator) validateObjectives(objectives []types.Objective) error {
	if len(objectives) == 0 {
		return fmt.Errorf("at least one objective is required")
	}

	totalWeight := 0.0
	for i, objective := range objectives {
		if err := v.validateObjective(&objective); err != nil {
			return fmt.Errorf("objective %d validation failed: %w", i, err)
		}
		totalWeight += objective.Weight
	}

	if totalWeight <= 0 {
		return fmt.Errorf("total weight must be greater than 0")
	}

	// Allow some tolerance for floating point precision
	if totalWeight < 0.99 || totalWeight > 1.01 {
		return fmt.Errorf("total weight should be approximately 1.0, got %f", totalWeight)
	}

	return nil
}

// validateObjective validates a single objective
func (v *Validator) validateObjective(objective *types.Objective) error {
	if objective == nil {
		return fmt.Errorf("objective cannot be nil")
	}

	if objective.Type == "" {
		return fmt.Errorf("objective type cannot be empty")
	}

	if objective.Weight <= 0 || objective.Weight > 1 {
		return fmt.Errorf("objective weight must be between 0 and 1, got %f", objective.Weight)
	}

	if objective.Target == "" {
		return fmt.Errorf("objective target cannot be empty")
	}

	return nil
}

// validateConstraints validates policy constraints
func (v *Validator) validateConstraints(constraints []types.Constraint) error {
	for i, constraint := range constraints {
		if err := v.validateConstraint(&constraint); err != nil {
			return fmt.Errorf("constraint %d validation failed: %w", i, err)
		}
	}
	return nil
}

// validateConstraint validates a single constraint
func (v *Validator) validateConstraint(constraint *types.Constraint) error {
	if constraint == nil {
		return fmt.Errorf("constraint cannot be nil")
	}

	if constraint.Type == "" {
		return fmt.Errorf("constraint type cannot be empty")
	}

	if constraint.Value == "" {
		return fmt.Errorf("constraint value cannot be empty")
	}

	return nil
}

// validateRules validates policy rules
func (v *Validator) validateRules(rules []types.Rule) error {
	for i, rule := range rules {
		if err := v.validateRule(&rule); err != nil {
			return fmt.Errorf("rule %d validation failed: %w", i, err)
		}
	}
	return nil
}

// validateRule validates a single rule
func (v *Validator) validateRule(rule *types.Rule) error {
	if rule == nil {
		return fmt.Errorf("rule cannot be nil")
	}

	if rule.Name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}

	if rule.Condition == "" {
		return fmt.Errorf("rule condition cannot be empty")
	}

	if rule.Action == "" {
		return fmt.Errorf("rule action cannot be empty")
	}

	return nil
}

// validateActions validates policy actions
func (v *Validator) validateActions(actions []types.Action) error {
	for i, action := range actions {
		if err := v.validateAction(&action); err != nil {
			return fmt.Errorf("action %d validation failed: %w", i, err)
		}
	}
	return nil
}

// validateAction validates a single action
func (v *Validator) validateAction(action *types.Action) error {
	if action == nil {
		return fmt.Errorf("action cannot be nil")
	}

	if action.Type == "" {
		return fmt.Errorf("action type cannot be empty")
	}

	return nil
}

// validateAutomationPolicy validates automation policy
func (v *Validator) validateAutomationPolicy(policy *types.Policy) error {
	// Add automation-specific validation logic here
	return nil
}

// validateWorkloadPriorityPolicy validates workload priority policy
func (v *Validator) validateWorkloadPriorityPolicy(policy *types.Policy) error {
	// Add workload priority-specific validation logic here
	return nil
}

// validateSecurityPolicy validates security policy
func (v *Validator) validateSecurityPolicy(policy *types.Policy) error {
	// Add security-specific validation logic here
	return nil
}

// validateResourceQuotaPolicy validates resource quota policy
func (v *Validator) validateResourceQuotaPolicy(policy *types.Policy) error {
	// Add resource quota-specific validation logic here
	return nil
}

// ValidateWorkload validates a workload
func (v *Validator) ValidateWorkload(workload *types.Workload) error {
	if workload == nil {
		return fmt.Errorf("workload cannot be nil")
	}

	if workload.ID == "" {
		return fmt.Errorf("workload ID cannot be empty")
	}

	if workload.Name == "" {
		return fmt.Errorf("workload name cannot be empty")
	}

	if workload.Type == "" {
		return fmt.Errorf("workload type cannot be empty")
	}

	if workload.Status == "" {
		return fmt.Errorf("workload status cannot be empty")
	}

	if err := v.validateLabels(workload.Labels); err != nil {
		return fmt.Errorf("workload labels validation failed: %w", err)
	}

	if err := v.validateAnnotations(workload.Annotations); err != nil {
		return fmt.Errorf("workload annotations validation failed: %w", err)
	}

	return nil
}

// ValidateAutomationRule validates an automation rule
func (v *Validator) ValidateAutomationRule(rule *types.AutomationRule) error {
	if rule == nil {
		return fmt.Errorf("automation rule cannot be nil")
	}

	if rule.ID == "" {
		return fmt.Errorf("automation rule ID cannot be empty")
	}

	if rule.Name == "" {
		return fmt.Errorf("automation rule name cannot be empty")
	}

	if rule.Type == "" {
		return fmt.Errorf("automation rule type cannot be empty")
	}

	if rule.Status == "" {
		return fmt.Errorf("automation rule status cannot be empty")
	}

	return nil
}

// ValidateExpression validates a policy expression
func (v *Validator) ValidateExpression(expression string) error {
	if expression == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	// Basic syntax validation
	if !strings.Contains(expression, "workload") && !strings.Contains(expression, "policy") {
		return fmt.Errorf("expression must reference workload or policy")
	}

	// Check for balanced parentheses
	if !v.isBalancedParentheses(expression) {
		return fmt.Errorf("expression has unbalanced parentheses")
	}

	return nil
}

// isBalancedParentheses checks if parentheses are balanced
func (v *Validator) isBalancedParentheses(expression string) bool {
	count := 0
	for _, char := range expression {
		if char == '(' {
			count++
		} else if char == ')' {
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

// ValidateTimeRange validates a time range
func (v *Validator) ValidateTimeRange(startTime, endTime time.Time) error {
	if startTime.IsZero() {
		return fmt.Errorf("start time cannot be zero")
	}

	if endTime.IsZero() {
		return fmt.Errorf("end time cannot be zero")
	}

	if startTime.After(endTime) {
		return fmt.Errorf("start time cannot be after end time")
	}

	return nil
}

// ValidatePercentage validates a percentage value
func (v *Validator) ValidatePercentage(value string) error {
	if value == "" {
		return fmt.Errorf("percentage value cannot be empty")
	}

	if !strings.HasSuffix(value, "%") {
		return fmt.Errorf("percentage value must end with %%")
	}

	// Extract numeric part
	numericPart := strings.TrimSuffix(value, "%")
	if numericPart == "" {
		return fmt.Errorf("percentage value must contain a numeric part")
	}

	// Basic validation - should be a number
	percentageRegex := regexp.MustCompile(`^\d+(\.\d+)?$`)
	if !percentageRegex.MatchString(numericPart) {
		return fmt.Errorf("percentage value must be a valid number")
	}

	return nil
}
