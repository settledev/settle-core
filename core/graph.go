package core

import (
	"fmt"
)

type Graph struct {
	nodes map[ResourceID]Resource
	edges map[ResourceID][]Dependency
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[ResourceID]Resource),
		edges: make(map[ResourceID][]Dependency),
	}
}

func (g *Graph) AddResource(resource Resource) error {
	if err := resource.Validate(); err != nil {
		return fmt.Errorf("invalid resource %s: %w", resource.GetID(), err)
	}

	g.nodes[resource.GetID()] = resource
	g.edges[resource.GetID()] = resource.GetDependencies()

	return nil
}

func (g *Graph) GetResource(id ResourceID) (Resource, bool) {
	resource, exists := g.nodes[id]
	if !exists {
		return nil, false
	}
	return resource, true
}

func (g *Graph) GetAllResources() []Resource {
	resources := make([]Resource, 0, len(g.nodes))
	for _, resource := range g.nodes {
		resources = append(resources, resource)
	}
	return resources
}

func (g *Graph) TopologicalSort() ([]ResourceID, error) {
	// Kahn's algorithm
	inDegree := make(map[ResourceID]int)

	// Initialize in-degree count
	for id := range g.nodes {
		inDegree[id] = 0
	}

	// Count incoming edges
	for _, deps := range g.edges {
		for _, dep := range deps {
			if dep.Required {
				inDegree[dep.Target]++
			}
		}
	}

	// Find nodes with no incoming edges
	queue := make([]ResourceID, 0)
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	result := make([]ResourceID, 0)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)


		for _, dep := range g.edges[current] {
			if dep.Required {
				inDegree[dep.Target]--
				if inDegree[dep.Target] == 0 {
					queue = append(queue, dep.Target)
				}
			}
		}
	}

	// Check for cycles
	if len(result) != len(g.nodes) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

func (g *Graph) ValidateDependencies() error {
	// Check for circular dependencies
	_, err := g.TopologicalSort()
	if err != nil {
		return err
	}

	// Validate layer dependencies
	for id, resource := range g.nodes {
		fromLayer := resource.GetLayer()

		for _, dep := range g.edges[id] {
			if dep.Required {
				targetResource, exists := g.nodes[dep.Target]
				if !exists {
					return fmt.Errorf("resource %s depends on non-existent resource %s", id, dep.Target)
				}

				toLayer := targetResource.GetLayer()
				if err := ValidateLayerDependency(fromLayer, toLayer); err != nil {
					return fmt.Errorf("resource %s: %w", id, err)
				}
			}
		}
	}

	return nil
}

func (g *Graph) GetDependents(id ResourceID) []ResourceID {
	dependents := make([]ResourceID, 0)

	for resourceID, deps := range g.edges {
		for _, dep := range deps {
			if dep.Target == id {
				dependents = append(dependents, resourceID)
				break
			}
		}
	}

	return dependents
}

func (g *Graph) GetDependencies(id ResourceID) []Dependency {
	return g.edges[id]
}

func (g *Graph) RemoveResource(id ResourceID) {
	delete(g.nodes, id)
	delete(g.edges, id)

	// Remove edges pointing to this resource
	for resourceID, deps := range g.edges {
		newDeps := make([]Dependency, 0)
		for _, dep := range deps {
			if dep.Target != id {
				newDeps = append(newDeps, dep)
			}
		}
		g.edges[resourceID] = newDeps
	}
}
