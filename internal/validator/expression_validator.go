package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/kcloud-opt/policy/internal/types"
)

// ExpressionValidator provides expression validation functionality
type ExpressionValidator struct {
	logger types.Logger
}

// NewExpressionValidator creates a new expression validator instance
func NewExpressionValidator(logger types.Logger) *ExpressionValidator {
	return &ExpressionValidator{
		logger: logger,
	}
}

// ValidateExpression validates a policy expression
func (ev *ExpressionValidator) ValidateExpression(expression string) error {
	if expression == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	// Parse the expression to check for syntax errors
	_, err := expr.Compile(expression, expr.Env(map[string]interface{}{
		"workload": map[string]interface{}{
			"id":        "",
			"name":      "",
			"type":      "",
			"status":    "",
			"namespace": "",
			"labels":    map[string]string{},
			"cpu": map[string]interface{}{
				"usage": 0.0,
				"limit": 0.0,
			},
			"memory": map[string]interface{}{
				"usage": 0.0,
				"limit": 0.0,
			},
			"storage": map[string]interface{}{
				"usage": 0.0,
				"limit": 0.0,
			},
		},
		"policy": map[string]interface{}{
			"id":       "",
			"name":     "",
			"type":     "",
			"status":   "",
			"priority": 0,
		},
		"cluster": map[string]interface{}{
			"resources": map[string]interface{}{
				"cpu":     0.0,
				"memory":  0.0,
				"storage": 0.0,
			},
		},
	}))
	if err != nil {
		return fmt.Errorf("expression syntax error: %w", err)
	}

	// Validate expression content
	if err := ev.validateExpressionContent(expression); err != nil {
		return fmt.Errorf("expression content validation failed: %w", err)
	}

	return nil
}

// validateExpressionContent validates the content of an expression
func (ev *ExpressionValidator) validateExpressionContent(expression string) error {
	// Check for dangerous functions or operations
	dangerousPatterns := []string{
		`\.(exec|system|eval|import|__)`,
		`\b(exec|system|eval|import|__)\s*\(`,
		`\b(os|sys|runtime)\s*\.`,
		`\b(panic|recover|defer)\s*\(`,
	}

	for _, pattern := range dangerousPatterns {
		matched, err := regexp.MatchString(pattern, expression)
		if err != nil {
			return fmt.Errorf("error checking dangerous pattern: %w", err)
		}
		if matched {
			return fmt.Errorf("expression contains potentially dangerous operation: %s", pattern)
		}
	}

	// Check for required context variables
	if !strings.Contains(expression, "workload") && !strings.Contains(expression, "policy") && !strings.Contains(expression, "cluster") {
		return fmt.Errorf("expression must reference at least one of: workload, policy, cluster")
	}

	// Check for balanced parentheses
	if !ev.isBalancedParentheses(expression) {
		return fmt.Errorf("expression has unbalanced parentheses")
	}

	// Check for balanced brackets
	if !ev.isBalancedBrackets(expression) {
		return fmt.Errorf("expression has unbalanced brackets")
	}

	return nil
}

// isBalancedParentheses checks if parentheses are balanced
func (ev *ExpressionValidator) isBalancedParentheses(expression string) bool {
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

// isBalancedBrackets checks if brackets are balanced
func (ev *ExpressionValidator) isBalancedBrackets(expression string) bool {
	count := 0
	for _, char := range expression {
		if char == '[' {
			count++
		} else if char == ']' {
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

// ValidateCondition validates a condition expression
func (ev *ExpressionValidator) ValidateCondition(condition string) error {
	if condition == "" {
		return fmt.Errorf("condition cannot be empty")
	}

	// Validate as expression
	if err := ev.ValidateExpression(condition); err != nil {
		return fmt.Errorf("condition validation failed: %w", err)
	}

	// Check that condition returns a boolean
	program, err := expr.Compile(condition, expr.Env(map[string]interface{}{
		"workload": map[string]interface{}{
			"cpu": map[string]interface{}{
				"usage": 0.0,
			},
			"memory": map[string]interface{}{
				"usage": 0.0,
			},
		},
	}))
	if err != nil {
		return fmt.Errorf("failed to compile condition: %w", err)
	}

	// Test with sample data
	result, err := expr.Run(program, map[string]interface{}{
		"workload": map[string]interface{}{
			"cpu": map[string]interface{}{
				"usage": 0.5,
			},
			"memory": map[string]interface{}{
				"usage": 0.6,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to evaluate condition: %w", err)
	}

	// Check if result is boolean
	if _, ok := result.(bool); !ok {
		return fmt.Errorf("condition must evaluate to a boolean value")
	}

	return nil
}

// ValidateRule validates a rule expression
func (ev *ExpressionValidator) ValidateRule(rule *types.Rule) error {
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

	// Validate condition
	if err := ev.ValidateCondition(rule.Condition); err != nil {
		return fmt.Errorf("rule condition validation failed: %w", err)
	}

	// Validate action (basic validation)
	if err := ev.validateAction(rule.Action); err != nil {
		return fmt.Errorf("rule action validation failed: %w", err)
	}

	return nil
}

// validateAction validates an action string
func (ev *ExpressionValidator) validateAction(action string) error {
	// Check for valid action types
	validActions := []string{
		"scale-up", "scale-down", "scale-workload",
		"reduce-cpu", "reduce-memory", "reduce-storage",
		"optimize-storage", "resource-adjustment",
		"notification", "alert", "log",
		"enable", "disable", "suspend",
	}

	actionLower := strings.ToLower(action)
	for _, validAction := range validActions {
		if strings.Contains(actionLower, validAction) {
			return nil
		}
	}

	// If no valid action found, check if it's a custom action
	if strings.Contains(action, "custom-") {
		return nil
	}

	return fmt.Errorf("invalid action: %s", action)
}

// ValidateTrigger validates a trigger expression
func (ev *ExpressionValidator) ValidateTrigger(trigger *types.Trigger) error {
	if trigger == nil {
		return fmt.Errorf("trigger cannot be nil")
	}

	if trigger.Type == "" {
		return fmt.Errorf("trigger type cannot be empty")
	}

	// Validate trigger type
	validTriggerTypes := []string{
		"event-based", "time-based", "threshold-based",
		"schedule-based", "condition-based",
	}

	triggerTypeLower := strings.ToLower(trigger.Type)
	valid := false
	for _, validType := range validTriggerTypes {
		if triggerTypeLower == validType {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid trigger type: %s", trigger.Type)
	}

	// Validate trigger-specific fields
	switch trigger.Type {
	case "event-based":
		if len(trigger.Events) == 0 {
			return fmt.Errorf("event-based trigger must have at least one event")
		}
	case "time-based":
		if trigger.Schedule == "" {
			return fmt.Errorf("time-based trigger must have a schedule")
		}
	case "threshold-based":
		if len(trigger.Metrics) == 0 {
			return fmt.Errorf("threshold-based trigger must have at least one metric")
		}
	}

	return nil
}

// ValidateAutomationRule validates an automation rule
func (ev *ExpressionValidator) ValidateAutomationRule(rule *types.AutomationRule) error {
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

	// Validate triggers
	if len(rule.Triggers) == 0 {
		return fmt.Errorf("automation rule must have at least one trigger")
	}

	for i, trigger := range rule.Triggers {
		if err := ev.ValidateTrigger(&trigger); err != nil {
			return fmt.Errorf("trigger %d validation failed: %w", i, err)
		}
	}

	// Validate conditions
	if len(rule.Conditions) == 0 {
		return fmt.Errorf("automation rule must have at least one condition")
	}

	for i, condition := range rule.Conditions {
		if err := ev.ValidateCondition(condition.Expression); err != nil {
			return fmt.Errorf("condition %d validation failed: %w", i, err)
		}
	}

	// Validate actions
	if len(rule.Actions) == 0 {
		return fmt.Errorf("automation rule must have at least one action")
	}

	for i, action := range rule.Actions {
		if err := ev.validateAction(action.Name); err != nil {
			return fmt.Errorf("action %d validation failed: %w", i, err)
		}
	}

	return nil
}
