package cmd

import (
	"context"
	"fmt"

	"github.com/settlectl/settle-core/common"
	"github.com/settlectl/settle-core/core"
	"github.com/settlectl/settle-core/inventory"
	"github.com/settlectl/settle-core/inventory/parser"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create units on hosts",
	Run: func(cmd *cobra.Command, args []string) {
		logger := inventory.NewLogger()
		logger.Info("Starting resource creation")


		hosts, err := parser.ParseHosts("hosts.stl")
		if err != nil {
			logger.Error(fmt.Sprintf("Error parsing hosts file: %v", err))
			return
		}
		logger.Info(fmt.Sprintf("Found %d hosts", len(hosts)))


		resourceFiles, err := findResourceFiles()
		if err != nil {
			logger.Error(fmt.Sprintf("Error finding resource files: %v", err))
			return
		}


		resourceParser := core.NewResourceParser()
		resourceParser.SetHosts(hosts)


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


		resources, err := resourceParser.ParseResources()
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating resources: %v", err))
			return
		}
		logger.Info(fmt.Sprintf("Created %d resources", len(resources)))


		graph := core.NewGraph()
		for _, resource := range resources {
			if err := graph.AddResource(resource); err != nil {
				logger.Error(fmt.Sprintf("Error adding resource %s to graph: %v", resource.GetID(), err))
				continue
			}
		}


		if err := graph.ValidateDependencies(); err != nil {
			logger.Error(fmt.Sprintf("Graph validation failed: %v", err))
			return
		}


		stateManager := core.NewStateManager(".settle/state.json", graph)
		if err := stateManager.LoadState(); err != nil {
			logger.Error(fmt.Sprintf("Error loading state: %v", err))
			return
		}


		planner := core.NewPlanner(graph, stateManager, logger)
		plan, err := planner.Plan()
		if err != nil {
			logger.Error(fmt.Sprintf("Error creating plan: %v", err))
			return
		}


		logger.Info("Execution Plan:")
		logger.Info(fmt.Sprintf("  Create: %d resources", plan.GetActionCount(core.ActionCreate)))
		logger.Info(fmt.Sprintf("  Update: %d resources", plan.GetActionCount(core.ActionUpdate)))
		logger.Info(fmt.Sprintf("  No-op: %d resources", plan.GetActionCount(core.ActionNoOp)))


		executor := core.NewExecutor(graph, stateManager, logger)
		executor.SetHosts(hosts)
		result, err := executor.Execute(context.Background(), plan)
		if err != nil {
			logger.Error(fmt.Sprintf("Execution failed: %v", err))
			return
		}


		logger.Info("Execution completed:")
		logger.Info(fmt.Sprintf("  Duration: %v", result.GetDuration()))
		logger.Info(fmt.Sprintf("  Success: %d", result.GetSuccessCount()))
		logger.Info(fmt.Sprintf("  Failed: %d", result.GetFailureCount()))
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
