package core

import (
	"fmt"
	"time"
	"github.com/settlectl/settle-core/inventory"
)

// Planner determines what actions need to be taken to reach desired state
type Planner struct {
	graph        *Graph
	stateManager *StateManager
	logger       *inventory.Logger
}

func NewPlanner(graph *Graph, stateManager *StateManager, logger *inventory.Logger) *Planner {
	return &Planner{
		graph:        graph,
		stateManager: stateManager,
		logger:       logger,
	}
}

// Plan creates an execution plan by comparing desired state with current state
func (p *Planner) Plan() (*Plan, error) {
	plan := &Plan{
		Actions:   make([]*Action, 0),
		CreatedAt: time.Now(),
		Graph:     p.graph,
	}

	// Get resources in dependency order
	resourceOrder, err := p.graph.TopologicalSort()
	if err != nil {
		return nil, fmt.Errorf("failed to sort resources: %w", err)
	}

	// Plan actions for each resource
	for _, resourceID := range resourceOrder {
		resource, exists := p.graph.GetResource(resourceID)
		if !exists {
			return nil, fmt.Errorf("resource %s not found in graph", resourceID)
		}

		action, err := p.planResource(resource)
		if err != nil {
			return nil, fmt.Errorf("failed to plan resource %s: %w", resourceID, err)
		}

		if action != nil {
			plan.Actions = append(plan.Actions, action)
		}
	}

	return plan, nil
}

// planResource determines what action (if any) is needed for a resource
func (p *Planner) planResource(resource Resource) (*Action, error) {
	// Check if resource exists in state
	currentState := p.stateManager.GetState(resource.GetID())

	// If resource doesn't exist in state, it needs to be created
	if currentState == nil {
		return &Action{
			ResourceID: resource.GetID(),
			Type:       ActionCreate,
			Changes:    []Change{},
			Metadata: map[string]interface{}{
				"reason": "resource not in state",
			},
		}, nil
	}

	// Check for configuration drift
	drifted, err := p.stateManager.DetectDrift(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to detect drift: %w", err)
	}

	if drifted {
		return &Action{
			ResourceID: resource.GetID(),
			Type:       ActionUpdate,
			Changes:    []Change{}, // TODO: Calculate actual changes
			Metadata: map[string]interface{}{
				"reason": "configuration drift detected",
			},
		}, nil
	}

	// Resource is up to date
	return &Action{
		ResourceID: resource.GetID(),
		Type:       ActionNoOp,
		Changes:    []Change{},
		Metadata: map[string]interface{}{
			"reason": "resource up to date",
		},
	}, nil
}

// Plan represents a complete execution plan
type Plan struct {
	Actions   []*Action `json:"actions"`
	CreatedAt time.Time `json:"created_at"`
	Graph     *Graph    `json:"graph"`
}

// ValidatePlan validates that the plan can be executed
func (p *Plan) ValidatePlan() error {
	// Check for circular dependencies
	_, err := p.Graph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("plan validation failed: %w", err)
	}

	// Validate dependencies
	err = p.Graph.ValidateDependencies()
	if err != nil {
		return fmt.Errorf("plan validation failed: %w", err)
	}

	return nil
}

// GetActionCount returns the count of actions by type
func (p *Plan) GetActionCount(actionType ActionType) int {
	count := 0
	for _, action := range p.Actions {
		if action.Type == actionType {
			count++
		}
	}
	return count
}

// GetActionsByType returns all actions of a specific type
func (p *Plan) GetActionsByType(actionType ActionType) []*Action {
	var actions []*Action
	for _, action := range p.Actions {
		if action.Type == actionType {
			actions = append(actions, action)
		}
	}
	return actions
}
