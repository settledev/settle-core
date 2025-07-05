package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/settlectl/settle-core/common"
	"github.com/settlectl/settle-core/core"
	"github.com/settlectl/settle-core/inventory"
	"github.com/settlectl/settle-core/inventory/parser"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean up resources",
	Run: func(cmd *cobra.Command, args []string) {
		logger := inventory.NewLogger()
		logger.Info("Starting resource cleanup")

		// Parse hosts
		hosts, err := parser.ParseHosts("hosts.stl")
		if err != nil {
			logger.Error(fmt.Sprintf("Error parsing hosts file: %v", err))
			return
		}
		logger.Info(fmt.Sprintf("Found %d hosts", len(hosts)))

		// Parse all resource files
		resourceFiles, err := findResourceFiles()
		if err != nil {
			logger.Error(fmt.Sprintf("Error finding resource files: %v", err))
			return
		}

		// Create resource parser and populate with data
		resourceParser := core.NewResourceParser()
		resourceParser.SetHosts(hosts)

		// Parse packages from all resource files
		var allPackages []common.Package
		for _, file := range resourceFiles {
			packages, err := parser.ParsePackages(file)
			if err != nil {
				logger.Error(fmt.Sprintf("Error parsing packages from %s: %v", file, err))
				continue
			}
			allPackages = append(allPackages, packages...)
		}
		resourceParser.SetPackages(allPackages)

		// Create resources using the parser
		resources, err := resourceParser.ParseResources()
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating resources: %v", err))
			return
		}
		logger.Info(fmt.Sprintf("Created %d resources for cleanup", len(resources)))

		// Create and populate the graph
		graph := core.NewGraph()
		for _, resource := range resources {
			if err := graph.AddResource(resource); err != nil {
				logger.Error(fmt.Sprintf("Error adding resource %s to graph: %v", resource.GetID(), err))
				continue
			}
		}

		// Validate the graph
		if err := graph.ValidateDependencies(); err != nil {
			logger.Error(fmt.Sprintf("Graph validation failed: %v", err))
			return
		}

		// Create state manager
		stateManager := core.NewStateManager(".settle/state.json", graph)
		if err := stateManager.LoadState(); err != nil {
			logger.Error(fmt.Sprintf("Error loading state: %v", err))
			return
		}

		// Create a cleanup plan (all resources marked for deletion)
		plan := &core.Plan{
			Actions:   make([]*core.Action, 0),
			CreatedAt: time.Now(),
			Graph:     graph,
		}

		// Mark all resources for deletion
		for _, resource := range resources {
			plan.Actions = append(plan.Actions, &core.Action{
				ResourceID: resource.GetID(),
				Type:       core.ActionDelete,
				Changes:    []core.Change{},
				Metadata: map[string]interface{}{
					"reason": "cleanup requested",
				},
			})
		}

		// Log plan summary
		logger.Info("Cleanup Plan:")
		logger.Info(fmt.Sprintf("  Delete: %d resources", len(plan.Actions)))

		// Create executor and execute the plan
		executor := core.NewExecutor(graph, stateManager, logger)
		executor.SetHosts(hosts)
		result, err := executor.Execute(context.Background(), plan)
		if err != nil {
			logger.Error(fmt.Sprintf("Cleanup failed: %v", err))
			return
		}

		// Log execution summary
		logger.Info("Cleanup completed:")
		logger.Info(fmt.Sprintf("  Duration: %v", result.GetDuration()))
		logger.Info(fmt.Sprintf("  Success: %d", result.GetSuccessCount()))
		logger.Info(fmt.Sprintf("  Failed: %d", result.GetFailureCount()))
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}

func detectResourceTypes(content string) []string {
	lines := strings.Split(content, "\n")
	resourceTypes := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "package ") && strings.Contains(line, "\"") {
			resourceTypes["package"] = true
		}
		if strings.HasPrefix(line, "service ") && strings.Contains(line, "\"") {
			resourceTypes["service"] = true
		}
		if strings.HasPrefix(line, "file ") && strings.Contains(line, "\"") {
			resourceTypes["file"] = true
		}
	}

	var types []string
	for resourceType := range resourceTypes {
		types = append(types, resourceType)
	}
	return types
}

func countResourceDeclarations(content string, resourceType string) int {
	lines := strings.Split(content, "\n")
	count := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, resourceType+" ") && strings.Contains(line, "\"") {
			count++
		}
	}
	return count
}

// Find all .stl files except hosts.stl
func findResourceFiles() ([]string, error) {
	files, err := filepath.Glob("*.stl")
	if err != nil {
		return nil, err
	}

	var resources []string
	for _, file := range files {
		if file == "hosts.stl" {
			continue // Skip hosts file
		}
		resources = append(resources, file)
	}
	return resources, nil
}
