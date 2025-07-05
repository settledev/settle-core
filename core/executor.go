package core

import (
	"context"
	"fmt"
	"time"

	"github.com/settlectl/settle-core/inventory"
	"github.com/settlectl/settle-core/common"
)

// Executor executes planned actions in dependency order
type Executor struct {
	graph        *Graph
	stateManager *StateManager
	logger       *inventory.Logger
	hosts        map[string]*common.Host // Map of host names to host objects
}

func NewExecutor(graph *Graph, stateManager *StateManager, logger *inventory.Logger) *Executor {
	return &Executor{
		graph:        graph,
		stateManager: stateManager,
		logger:       logger,
		hosts:        make(map[string]*common.Host),
	}
}

// SetHosts sets the hosts available for execution
func (e *Executor) SetHosts(hosts []common.Host) {
	e.hosts = make(map[string]*common.Host)
	for i := range hosts {
		e.hosts[hosts[i].Name] = &hosts[i]
	}
}

// Execute runs a complete execution plan
func (e *Executor) Execute(ctx context.Context, plan *Plan) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Plan:      plan,
		StartedAt: time.Now(),
		Actions:   make([]*ExecutionAction, 0),
	}

	// Validate the plan before execution
	if err := plan.ValidatePlan(); err != nil {
		return nil, fmt.Errorf("plan validation failed: %w", err)
	}

	e.logger.Info("Starting execution of plan")
	e.logger.Info(fmt.Sprintf("Plan contains %d actions", len(plan.Actions)))

	// Execute actions in order
	for i, action := range plan.Actions {
		e.logger.Info(fmt.Sprintf("Executing action %d/%d: %s", i+1, len(plan.Actions), action.ResourceID))

		execAction, err := e.executeAction(ctx, action)
		if err != nil {
			result.FailedAt = time.Now()
			result.Error = err
			return result, fmt.Errorf("execution failed at action %s: %w", action.ResourceID, err)
		}

		result.Actions = append(result.Actions, execAction)
	}

	result.CompletedAt = time.Now()
	result.Success = true
	e.logger.Info("Execution completed successfully")

	return result, nil
}

// executeAction executes a single action
func (e *Executor) executeAction(ctx context.Context, action *Action) (*ExecutionAction, error) {
	execAction := &ExecutionAction{
		Action:    action,
		StartedAt: time.Now(),
	}

	// Get the resource
	resource, exists := e.graph.GetResource(action.ResourceID)
	if !exists {
		execAction.FailedAt = time.Now()
		execAction.Error = fmt.Errorf("resource %s not found", action.ResourceID)
		return execAction, execAction.Error
	}

	// Create context for the resource
	resourceCtx := e.createResourceContext(resource)

	// Execute based on action type
	var err error
	switch action.Type {
	case ActionCreate:
		err = resource.Apply(resourceCtx)
	case ActionUpdate:
		err = resource.Apply(resourceCtx)
	case ActionDelete:
		err = resource.Destroy(resourceCtx)
	case ActionNoOp:
		e.logger.Info(fmt.Sprintf("Skipping %s (no-op)", action.ResourceID))
		execAction.CompletedAt = time.Now()
		return execAction, nil
	default:
		err = fmt.Errorf("unknown action type: %s", action.Type)
	}

	if err != nil {
		execAction.FailedAt = time.Now()
		execAction.Error = err

		// Mark resource as failed in state
		e.stateManager.MarkFailed(resource, err.Error())

		return execAction, fmt.Errorf("action failed: %w", err)
	}

	// Mark resource as applied in state
	err = e.stateManager.MarkApplied(resource)
	if err != nil {
		execAction.FailedAt = time.Now()
		execAction.Error = err
		return execAction, fmt.Errorf("failed to mark resource as applied: %w", err)
	}

	execAction.CompletedAt = time.Now()
	e.logger.Info(fmt.Sprintf("Successfully executed %s", action.ResourceID))

	return execAction, nil
}

// createResourceContext creates a context for resource execution
func (e *Executor) createResourceContext(resource Resource) *inventory.Context {
	// Create a basic context
	ctx := &inventory.Context{
		Logger: e.logger,
	}

	// For host resources, set the host
	if hostResource, ok := resource.(*HostResource); ok {
		ctx.SetHost(&hostResource.Host)
	}

	// For package resources, try to find the host they should be installed on
	// This is a simplified approach - in a real system, you'd have explicit dependencies
	if _, ok := resource.(*PackageResource); ok {
		// For now, use the first available host
		// In a real system, you'd determine this from dependencies
		for _, host := range e.hosts {
			ctx.SetHost(host)
			break
		}
	}

	return ctx
}

// ExecutionResult represents the result of an execution
type ExecutionResult struct {
	Plan        *Plan              `json:"plan"`
	StartedAt   time.Time          `json:"started_at"`
	CompletedAt time.Time          `json:"completed_at,omitempty"`
	FailedAt    time.Time          `json:"failed_at,omitempty"`
	Success     bool               `json:"success"`
	Error       error              `json:"error,omitempty"`
	Actions     []*ExecutionAction `json:"actions"`
}

// ExecutionAction represents the result of executing a single action
type ExecutionAction struct {
	Action      *Action   `json:"action"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	FailedAt    time.Time `json:"failed_at,omitempty"`
	Error       error     `json:"error,omitempty"`
}

// GetDuration returns the total execution duration
func (r *ExecutionResult) GetDuration() time.Duration {
	if r.Success {
		return r.CompletedAt.Sub(r.StartedAt)
	}
	return r.FailedAt.Sub(r.StartedAt)
}

// GetSuccessCount returns the number of successful actions
func (r *ExecutionResult) GetSuccessCount() int {
	count := 0
	for _, action := range r.Actions {
		if action.Error == nil {
			count++
		}
	}
	return count
}

// GetFailureCount returns the number of failed actions
func (r *ExecutionResult) GetFailureCount() int {
	count := 0
	for _, action := range r.Actions {
		if action.Error != nil {
			count++
		}
	}
	return count
}
