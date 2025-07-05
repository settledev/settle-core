package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type StateManager struct {
	stateFile string
	state     map[ResourceID]*ResourceState
	graph     *Graph
}

func NewStateManager(stateFile string, graph *Graph) *StateManager {
	return &StateManager{
		stateFile: stateFile,
		state:     make(map[ResourceID]*ResourceState),
		graph:     graph,
	}
}

func (s *StateManager) LoadState() error {
	if _, err := os.Stat(s.stateFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(s.stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var stateData map[ResourceID]*ResourceState
	if err := json.Unmarshal(data, &stateData); err != nil {
		return fmt.Errorf("failed to unmarshal state file: %w", err)
	}

	s.state = stateData
	return nil
}

func (s *StateManager) SaveState() error {
	dir := filepath.Dir(s.stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(s.stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func (s *StateManager) GetState(id ResourceID) *ResourceState {
	state, exists := s.state[id]
	if !exists {
		return nil
	}
	return state
}

func (s *StateManager) SetState(id ResourceID, state *ResourceState) {
	s.state[id] = state
}

func (s *StateManager) RemoveState(id ResourceID) {
	delete(s.state, id)
}

func (s *StateManager) GetAllStates() map[ResourceID]*ResourceState {
	result := make(map[ResourceID]*ResourceState)
	for id, state := range s.state {
		result[id] = state
	}
	return result
}

func (s *StateManager) DetectDrift(resource Resource) (bool, error) {
	currentState := s.GetState(resource.GetID())
	if currentState == nil {
		// Resource not in state, consider it drifted
		return true, nil
	}

	currentConfig := resource.GetConfig()
	lastConfig, exists := currentState.Metadata["config"]
	if !exists {
		return true, nil
	}

	configBytes, err := json.Marshal(currentConfig)
	if err != nil {
		return false, fmt.Errorf("failed to marshal current config: %w", err)
	}

	lastConfigBytes, err := json.Marshal(lastConfig)
	if err != nil {
		return false, fmt.Errorf("failed to marshal last config: %w", err)
	}

	return string(configBytes) != string(lastConfigBytes), nil
}

func (s *StateManager) MarkApplied(resource Resource) error {
	config := resource.GetConfig()
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	state := &ResourceState{
		Status:      StateApplied,
		LastApplied: time.Now(),
		Checksum:    string(configBytes),
		Metadata: map[string]interface{}{
			"config": config,
		},
	}

	s.SetState(resource.GetID(), state)
	return s.SaveState()
}

func (s *StateManager) MarkFailed(resource Resource, errorMsg string) error {
	state := &ResourceState{
		Status:      StateFailed,
		LastApplied: time.Now(),
		Metadata: map[string]interface{}{
			"error": errorMsg,
		},
	}

	s.SetState(resource.GetID(), state)
	return s.SaveState()
}

func (s *StateManager) GetResourcesByStatus(status StateStatus) []ResourceID {
	var result []ResourceID
	for id, state := range s.state {
		if state.Status == status {
			result = append(result, id)
		}
	}
	return result
}

// Cleanup removes states for resources that no longer exist in the graph
func (s *StateManager) Cleanup() error {
	graphResources := make(map[ResourceID]bool)
	for _, resource := range s.graph.GetAllResources() {
		graphResources[resource.GetID()] = true
	}

	var toRemove []ResourceID
	for id := range s.state {
		if !graphResources[id] {
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		s.RemoveState(id)
	}

	if len(toRemove) > 0 {
		return s.SaveState()
	}
	return nil
}
