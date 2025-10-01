package automation

import (
	"context"
	"time"
)

// AutomationEngine defines the interface for automation engine
type AutomationEngine interface {
	// Start starts the automation engine
	Start(ctx context.Context) error

	// Stop stops the automation engine
	Stop(ctx context.Context) error

	// RegisterRule registers an automation rule
	RegisterRule(ctx context.Context, rule *AutomationRule) error

	// UnregisterRule unregisters an automation rule
	UnregisterRule(ctx context.Context, ruleID string) error

	// TriggerRule manually triggers a rule
	TriggerRule(ctx context.Context, ruleID string, context map[string]interface{}) error

	// GetRuleStatus gets the status of a rule
	GetRuleStatus(ctx context.Context, ruleID string) (*RuleStatus, error)

	// GetRules returns all registered rules
	GetRules(ctx context.Context) ([]*AutomationRule, error)

	// Health checks the health of the automation engine
	Health(ctx context.Context) error
}

// AutomationRule represents an automation rule
type AutomationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Conditions  []*Condition           `json:"conditions"`
	Actions     []*Action              `json:"actions"`
	Schedule    *Schedule              `json:"schedule,omitempty"`
	Triggers    []*Trigger             `json:"triggers"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// Condition represents a condition for automation
type Condition struct {
	Field    string                 `json:"field"`
	Operator string                 `json:"operator"`
	Value    interface{}            `json:"value"`
	Duration *time.Duration         `json:"duration,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Action represents an automation action
type Action struct {
	Type       string                 `json:"type"`
	Target     string                 `json:"target,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Timeout    *time.Duration         `json:"timeout,omitempty"`
	Retry      *RetryConfig           `json:"retry,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Schedule represents a schedule for time-based automation
type Schedule struct {
	Cron      string    `json:"cron,omitempty"`
	Interval  string    `json:"interval,omitempty"`
	Timezone  string    `json:"timezone,omitempty"`
	StartTime time.Time `json:"startTime,omitempty"`
	EndTime   time.Time `json:"endTime,omitempty"`
}

// Trigger represents an event trigger
type Trigger struct {
	Type     string                 `json:"type"`
	Event    string                 `json:"event"`
	Filters  map[string]interface{} `json:"filters,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxRetries int           `json:"maxRetries"`
	Interval   time.Duration `json:"interval"`
	Backoff    string        `json:"backoff,omitempty"`
}

// RuleStatus represents the status of an automation rule
type RuleStatus struct {
	RuleID         string                 `json:"ruleId"`
	Name           string                 `json:"name"`
	Status         RuleExecutionStatus    `json:"status"`
	LastExecuted   *time.Time             `json:"lastExecuted,omitempty"`
	NextExecution  *time.Time             `json:"nextExecution,omitempty"`
	ExecutionCount int64                  `json:"executionCount"`
	SuccessCount   int64                  `json:"successCount"`
	FailureCount   int64                  `json:"failureCount"`
	LastError      string                 `json:"lastError,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// RuleExecutionStatus represents the execution status of a rule
type RuleExecutionStatus string

const (
	RuleStatusActive   RuleExecutionStatus = "active"
	RuleStatusInactive RuleExecutionStatus = "inactive"
	RuleStatusRunning  RuleExecutionStatus = "running"
	RuleStatusFailed   RuleExecutionStatus = "failed"
	RuleStatusDisabled RuleExecutionStatus = "disabled"
)

// Event represents an automation event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EventHandler defines the interface for handling events
type EventHandler interface {
	// HandleEvent handles an automation event
	HandleEvent(ctx context.Context, event *Event) error

	// CanHandle checks if this handler can handle the event type
	CanHandle(eventType string) bool

	// Health checks the health of the event handler
	Health(ctx context.Context) error
}

// RuleExecutor defines the interface for executing automation rules
type RuleExecutor interface {
	// ExecuteRule executes an automation rule
	ExecuteRule(ctx context.Context, rule *AutomationRule, context map[string]interface{}) (*ExecutionResult, error)

	// ValidateRule validates an automation rule
	ValidateRule(ctx context.Context, rule *AutomationRule) error

	// Health checks the health of the rule executor
	Health(ctx context.Context) error
}

// ExecutionResult represents the result of rule execution
type ExecutionResult struct {
	RuleID    string                 `json:"ruleId"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Actions   []*ActionResult        `json:"actions,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ActionResult represents the result of an action execution
type ActionResult struct {
	ActionType string                 `json:"actionType"`
	Success    bool                   `json:"success"`
	Message    string                 `json:"message"`
	Duration   time.Duration          `json:"duration"`
	Timestamp  time.Time              `json:"timestamp"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Error      string                 `json:"error,omitempty"`
	RetryCount int                    `json:"retryCount"`
}

// ConditionEvaluator defines the interface for evaluating conditions
type ConditionEvaluator interface {
	// EvaluateCondition evaluates a condition against context
	EvaluateCondition(ctx context.Context, condition *Condition, context map[string]interface{}) (bool, error)

	// EvaluateConditions evaluates multiple conditions
	EvaluateConditions(ctx context.Context, conditions []*Condition, context map[string]interface{}) (bool, error)

	// Health checks the health of the condition evaluator
	Health(ctx context.Context) error
}

// ActionExecutor defines the interface for executing actions
type ActionExecutor interface {
	// ExecuteAction executes an action
	ExecuteAction(ctx context.Context, action *Action) (*ActionResult, error)

	// CanExecute checks if this executor can handle the action type
	CanExecute(actionType string) bool

	// Health checks the health of the action executor
	Health(ctx context.Context) error
}

// Scheduler defines the interface for scheduling automation rules
type Scheduler interface {
	// ScheduleRule schedules a rule for execution
	ScheduleRule(ctx context.Context, rule *AutomationRule) error

	// UnscheduleRule unschedules a rule
	UnscheduleRule(ctx context.Context, ruleID string) error

	// GetScheduledRules returns all scheduled rules
	GetScheduledRules(ctx context.Context) ([]*AutomationRule, error)

	// Health checks the health of the scheduler
	Health(ctx context.Context) error
}

// Common automation event types
const (
	EventTypeWorkloadCreated   = "workload.created"
	EventTypeWorkloadUpdated   = "workload.updated"
	EventTypeWorkloadDeleted   = "workload.deleted"
	EventTypeWorkloadCompleted = "workload.completed"
	EventTypeWorkloadFailed    = "workload.failed"

	EventTypePolicyCreated = "policy.created"
	EventTypePolicyUpdated = "policy.updated"
	EventTypePolicyDeleted = "policy.deleted"

	EventTypeDecisionCreated   = "decision.created"
	EventTypeDecisionCompleted = "decision.completed"
	EventTypeDecisionFailed    = "decision.failed"

	EventTypeSchedule = "schedule"
	EventTypeManual   = "manual"
)

// Common automation action types
const (
	ActionTypeNotify     = "notify"
	ActionTypeScale      = "scale"
	ActionTypeMigrate    = "migrate"
	ActionTypeTerminate  = "terminate"
	ActionTypeSuspend    = "suspend"
	ActionTypeResume     = "resume"
	ActionTypeUpdate     = "update"
	ActionTypeCreate     = "create"
	ActionTypeDelete     = "delete"
	ActionTypeSchedule   = "schedule"
	ActionTypeReschedule = "reschedule"
	ActionTypeOptimize   = "optimize"
)

// Common condition operators
const (
	OperatorEquals             = "equals"
	OperatorNotEquals          = "not_equals"
	OperatorGreaterThan        = "greater_than"
	OperatorLessThan           = "less_than"
	OperatorGreaterThanOrEqual = "greater_than_or_equal"
	OperatorLessThanOrEqual    = "less_than_or_equal"
	OperatorContains           = "contains"
	OperatorNotContains        = "not_contains"
	OperatorStartsWith         = "starts_with"
	OperatorEndsWith           = "ends_with"
	OperatorRegex              = "regex"
	OperatorIn                 = "in"
	OperatorNotIn              = "not_in"
)
